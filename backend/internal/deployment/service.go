package deployment

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/system"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"

	"github.com/senvanda/backend/internal/cicd"
	"github.com/senvanda/backend/internal/container"
	"github.com/senvanda/backend/internal/git"
)

type service struct {
	app        core.App
	containers container.Service
	git        git.Service
	cicd       cicd.Service
}

func NewService(app core.App, containerSvc container.Service, gitSvc git.Service, cicdSvc cicd.Service) Service {
	return &service{
		app:        app,
		containers: containerSvc,
		git:        gitSvc,
		cicd:       cicdSvc,
	}
}

func (s *service) GetDockerInfo(ctx context.Context) (system.Info, error) {
	// For info, we might still need a direct call or add to container interface
	// I'll keep it simple for now and assume container service can provide it if needed
	// But system.Info is very specific.
	return system.Info{}, nil // Placeholder or extend container service
}

func (s *service) GetProjectsWithStatus(ctx context.Context) ([]ProjectStatus, error) {
	// 1. Fetch current Managed Projects from DB
	records, err := s.app.Dao().FindRecordsByFilter("projects", "id != ''", "-created", 200, 0, nil)
	if err != nil {
		return nil, err
	}

	// 2. Identify Orphans (Containers with 'senvanda-' prefix but no DB record)
	containers, _ := s.containers.ListContainers(ctx, true)
	dbMap := make(map[string]bool)
	for _, r := range records {
		dbMap[r.GetString("containerId")] = true
		dbMap["senvanda-"+r.GetString("name")] = true
		dbMap[r.GetString("name")] = true
	}

	for _, c := range containers {
		fullName := ""
		if len(c.Names) > 0 {
			fullName = strings.TrimPrefix(c.Names[0], "/")
		}

		// If it looks like a Senvanda container but isn't in our DB -> RECOVER IT
		if strings.HasPrefix(fullName, "senvanda-") && !dbMap[fullName] && !dbMap[c.ID] {
			fmt.Printf("[RECOVERY] Orphaned Senvanda container found: %s. Auto-healing record...\n", fullName)

			// Get default system user
			user, _ := s.FindFirstUser(ctx)

			// Build Record
			collection, _ := s.app.Dao().FindCollectionByNameOrId("projects")
			rec := models.NewRecord(collection)

			cleanName := strings.TrimPrefix(fullName, "senvanda-")
			rec.Set("name", cleanName)
			rec.Set("user", user.Id)
			rec.Set("status", "running")
			rec.Set("containerId", c.ID)
			rec.Set("image", c.Image)
			rec.Set("webhookToken", s.cicd.GenerateWebhookToken())

			// Try to find a port if exposed
			if len(c.Ports) > 0 {
				rec.Set("port", int(c.Ports[0].PublicPort))
			} else {
				port, _ := s.findAvailablePort()
				rec.Set("port", port)
			}

			if err := s.app.Dao().SaveRecord(rec); err == nil {
				records = append([]*models.Record{rec}, records...)
			}
		}
	}

	// 3. Prepare Final Status List
	var results []ProjectStatus
	for _, r := range records {
		name := r.GetString("name")
		cid := r.GetString("containerId")
		status := "stopped"
		state := ""

		if r.GetString("status") == "draft" {
			status = "draft"
		} else {
			// DEVOPS: Smart Container Detection
			var cJSON types.ContainerJSON
			var err error

			if cid != "" {
				cJSON, err = s.containers.InspectContainer(ctx, cid)
			} else {
				// Try Prefixed then Direct
				cJSON, err = s.containers.InspectContainer(ctx, "senvanda-"+name)
				if err != nil {
					cJSON, err = s.containers.InspectContainer(ctx, name)
				}
			}

			if err == nil {
				state = cJSON.State.Status
				status = "running"
				if !cJSON.State.Running {
					status = cJSON.State.Status
				}
			} else {
				status = "missing"
			}
		}

		results = append(results, ProjectStatus{
			ID:       r.Id,
			Name:     name,
			Port:     r.GetInt("port"),
			DBStatus: r.GetString("status"),
			Status:   status,
			State:    state,
			Created:  r.Created,
			Image:    r.GetString("image"),
			RepoUrl:  r.GetString("repoUrl"),
		})
	}

	return results, nil
}

func (s *service) ActionProject(ctx context.Context, projectID string, action string) error {
	record, err := s.app.Dao().FindRecordById("projects", projectID)
	if err != nil {
		return err
	}

	containerName := "senvanda-" + record.GetString("name")

	switch action {
	case "start":
		return s.containers.StartContainer(ctx, containerName)
	case "stop":
		return s.containers.StopContainer(ctx, containerName)
	case "restart":
		return s.containers.RestartContainer(ctx, containerName)
	case "redeploy":
		// Cleanup
		_ = s.containers.RemoveContainer(ctx, containerName)

		// Prepare Config
		port := record.GetInt("port")
		name := record.GetString("name")
		image := record.GetString("image")
		repoUrl := record.GetString("repoUrl")

		if image == "custom-build" && repoUrl != "" {
			// DEVOPS: Build from Source
			tempPath := filepath.Join(os.TempDir(), "senvanda-build-"+name)
			_ = os.RemoveAll(tempPath)

			// Clone (We can repurpose git service or use exec)
			cmd := exec.Command("git", "clone", "--depth", "1", repoUrl, tempPath)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to clone for build: %v", err)
			}
			defer os.RemoveAll(tempPath)

			// Build
			tag := "senvanda/project-" + name + ":latest"
			if err := s.containers.BuildImage(ctx, tempPath, tag); err != nil {
				return fmt.Errorf("build failed: %v", err)
			}
			image = tag
		}

		if image == "" || image == "custom-build" {
			image = "nginx:alpine"
		}

		var envs []string
		var cpu, memory string
		settings := record.Get("settings")

		if data, ok := settings.(map[string]interface{}); ok {
			if envList, ok := data["envVars"].([]interface{}); ok {
				for _, e := range envList {
					if kv, ok := e.(map[string]interface{}); ok {
						key := fmt.Sprintf("%v", kv["key"])
						val := fmt.Sprintf("%v", kv["value"])
						if key != "" {
							envs = append(envs, fmt.Sprintf("%s=%s", key, val))
						}
					}
				}
			}
			if res, ok := data["resources"].(map[string]interface{}); ok {
				cpu = fmt.Sprintf("%v", res["cpu"])
				memory = fmt.Sprintf("%v", res["memory"])
			}
		}

		// Domain Logic
		domain := fmt.Sprintf("%s.senvanda.local", name)
		if data, ok := settings.(map[string]interface{}); ok {
			if d, ok := data["domain"].(string); ok && d != "" {
				domain = d
			}
		}

		containerCfg := &container.Config{
			Name:  containerName,
			Image: image,
			Env:   envs,
			Ports: map[string]string{"80/tcp": strconv.Itoa(port)},
			Labels: map[string]string{
				"senvanda.project":    name,
				"senvanda.redeployed": time.Now().Format(time.RFC3339),
				"caddy":               domain,
				"caddy.reverse_proxy": "{{upstreams 80}}",
			},
		}
		containerCfg.Resources.CPU = cpu
		containerCfg.Resources.Memory = memory

		id, err := s.containers.CreateContainer(ctx, containerCfg)

		if err != nil {
			record.Set("status", "failed")
			s.app.Dao().SaveRecord(record)
			return err
		}

		if err := s.containers.StartContainer(ctx, containerName); err != nil {
			record.Set("status", "failed")
			s.app.Dao().SaveRecord(record)
			return err
		}

		record.Set("containerId", id)
		record.Set("status", "running")
		s.app.Dao().SaveRecord(record)
		return nil
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (s *service) ActionProjectByToken(ctx context.Context, token string, action string) error {
	record, err := s.app.Dao().FindFirstRecordByData("projects", "webhookToken", token)
	if err != nil {
		return fmt.Errorf("invalid token")
	}
	return s.ActionProject(ctx, record.Id, action)
}

func (s *service) CreateProject(ctx context.Context, req CreateProjectReq, user *models.Record) (*models.Record, error) {
	// 1. Validation & Hygiene
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || strings.ToLower(req.Name) == "untitled" || strings.ToLower(req.Name) == "untitled project" {
		return nil, fmt.Errorf("invalid project name: '%s'. please provide a descriptive name.", req.Name)
	}

	// 1b. Port allocation
	port := req.Port
	if port == 0 {
		var err error
		port, err = s.findAvailablePort()
		if err != nil {
			return nil, err
		}
	}

	// 2. DB Record
	collection, err := s.app.Dao().FindCollectionByNameOrId("projects")
	if err != nil {
		return nil, err
	}

	record := models.NewRecord(collection)
	record.Set("name", req.Name)
	record.Set("port", port)
	record.Set("user", user.Id)
	record.Set("repoUrl", req.RepoUrl)
	record.Set("framework", req.Framework)
	record.Set("image", req.Image)
	record.Set("webhookToken", s.cicd.GenerateWebhookToken())
	record.Set("settings", req.Settings)
	record.Set("status", "draft")
	if !req.IsDraft {
		record.Set("status", "building")
	}

	if err := s.app.Dao().SaveRecord(record); err != nil {
		return nil, err
	}

	if req.IsDraft {
		return record, nil
	}

	// 3. Deployment
	err = s.ActionProject(ctx, record.Id, "redeploy")
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *service) ScanGitRepository(ctx context.Context, repoUrl string) (*ScanResult, error) {
	res, err := s.git.ScanRepository(ctx, repoUrl)
	if err != nil {
		return nil, err
	}

	var envs []EnvVar
	for _, e := range res.EnvVars {
		envs = append(envs, EnvVar{Key: e.Key, Value: e.Value})
	}

	return &ScanResult{
		Framework:     res.Framework,
		Version:       res.Version,
		StartCommand:  res.StartCommand,
		Port:          res.Port,
		Image:         res.Image,
		EnvVars:       envs,
		TracingLogs:   res.TracingLogs,
		SecurityHints: res.SecurityHints,
		DevOpsProfile: res.DevOpsProfile,
		Name:          extractNameFromUrl(repoUrl),
		Domain:        fmt.Sprintf("%s.senvanda.local", extractNameFromUrl(repoUrl)),
	}, nil
}

func (s *service) DiscoverLegacy(ctx context.Context) ([]LegacyApp, error) {
	// 1. Get all containers from Docker
	containers, err := s.containers.ListContainers(ctx, true)
	if err != nil {
		return nil, err
	}

	// 2. Get all project records from DB to check for membership
	records, err := s.app.Dao().FindRecordsByFilter("projects", "id != ''", "", 1000, 0, nil)
	if err != nil {
		return nil, err
	}

	managedMap := make(map[string]bool)
	for _, r := range records {
		managedMap[r.GetString("containerId")] = true
		managedMap["senvanda-"+r.GetString("name")] = true
		managedMap[r.GetString("name")] = true
	}

	var apps []LegacyApp
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		// Skip internal Senvanda infrastructure
		if strings.HasPrefix(name, "infrastructure-") || name == "pocketbase" || name == "caddy" {
			continue
		}

		// IS IT MANAGED? Check ID and Names
		isManaged := managedMap[c.ID] || managedMap[name] || managedMap[strings.TrimPrefix(name, "senvanda-")]

		if !isManaged {
			var ports []int
			for _, p := range c.Ports {
				if p.PublicPort != 0 {
					ports = append(ports, int(p.PublicPort))
				}
			}

			apps = append(apps, LegacyApp{
				ID:    c.ID,
				Name:  name,
				Image: c.Image,
				State: c.State,
				Port:  ports,
			})
		}
	}
	return apps, nil
}
func (s *service) GetProjectLogs(ctx context.Context, projectID string) (string, error) {
	record, err := s.app.Dao().FindRecordById("projects", projectID)
	if err != nil {
		return "", err
	}

	containerName := "senvanda-" + record.GetString("name")
	return s.containers.GetContainerLogs(ctx, containerName)
}

func (s *service) AdoptProject(ctx context.Context, containerID string, userID string) (*models.Record, error) {
	details, err := s.containers.GetContainerDetails(ctx, containerID)
	if err != nil {
		return nil, err
	}

	collection, _ := s.app.Dao().FindCollectionByNameOrId("projects")
	record := models.NewRecord(collection)

	name := strings.TrimPrefix(details.Name, "/")
	record.Set("name", name)
	record.Set("user", userID)
	record.Set("status", "running")
	record.Set("image", details.Image)
	record.Set("containerId", containerID)

	if len(details.Ports) > 0 {
		record.Set("port", details.Ports[0])
	}

	if err := s.app.Dao().SaveRecord(record); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *service) PruneMissingProjects(ctx context.Context) (int, error) {
	records, err := s.app.Dao().FindRecordsByFilter("projects", "id != ''", "", 2000, 0, nil)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, r := range records {
		name := strings.TrimSpace(r.GetString("name"))
		cid := r.GetString("containerId")

		// 1. ANONYMOUS PRUNING (Aggressive)
		if name == "" || strings.ToLower(name) == "untitled" || strings.ToLower(name) == "untitled project" {
			fmt.Printf("[PRUNE] Removing anonymous zombie: %s\n", r.Id)
			if err := s.app.Dao().DeleteRecord(r); err == nil {
				count++
			}
			continue
		}

		// 2. DRAFT PRUNING (Cleanup ancient drafts)
		if r.GetString("status") == "draft" {
			createdAt := r.GetDateTime("created").Time()
			if time.Since(createdAt) > 24*time.Hour {
				_ = s.app.Dao().DeleteRecord(r)
				count++
			}
			continue
		}

		// 3. DEVOPS GRACE PERIOD: Don't prune if updated in last 2 minutes
		if time.Since(r.GetDateTime("updated").Time()) < 2*time.Minute {
			continue
		}

		var exists bool
		var err error

		if cid != "" {
			exists, err = s.containers.ContainerExists(ctx, cid)
		} else {
			// Legacy Fallback
			prefixed := name
			if !strings.HasPrefix(name, "senvanda-") {
				prefixed = "senvanda-" + name
			}
			e1, _ := s.containers.ContainerExists(ctx, prefixed)
			e2, _ := s.containers.ContainerExists(ctx, name)
			exists = e1 || e2
		}

		if err != nil {
			fmt.Printf("[PRUNE] Docker connection error for %s: %v\n", name, err)
			continue
		}

		if !exists {
			fmt.Printf("[PRUNE] DELETING GHOST: %s (ID: %s / CID: %s)\n", name, r.Id, cid)
			if err := s.app.Dao().DeleteRecord(r); err == nil {
				count++
			}
		}
	}
	return count, nil
}

func (s *service) findAvailablePort() (int, error) {
	reservedPorts := map[int]bool{22: true, 80: true, 443: true, 3000: true, 8090: true, 9443: true}
	usedInDB := make(map[int]bool)
	records, _ := s.app.Dao().FindRecordsByFilter("projects", "port > 0", "", 1000, 0, nil)
	for _, r := range records {
		usedInDB[r.GetInt("port")] = true
	}

	for p := 10000; p < 20000; p++ {
		if reservedPorts[p] || usedInDB[p] {
			continue
		}
		l, err := net.Listen("tcp", ":"+strconv.Itoa(p))
		if err == nil {
			l.Close()
			return p, nil
		}
	}
	return 0, fmt.Errorf("no port available")
}

func (s *service) FindFirstUser(ctx context.Context) (*models.Record, error) {
	records, err := s.app.Dao().FindRecordsByFilter("users", "id != ''", "-created", 1, 0, nil)
	if err == nil && len(records) > 0 {
		return records[0], nil
	}

	collection, _ := s.app.Dao().FindCollectionByNameOrId("users")
	record := models.NewRecord(collection)
	record.Set("email", "system@senvanda.local")
	record.SetPassword("1234567890")
	s.app.Dao().SaveRecord(record)
	return record, nil
}

func extractNameFromUrl(url string) string {
	parts := strings.Split(url, "/")
	last := parts[len(parts)-1]
	return strings.TrimSuffix(last, ".git")
}
