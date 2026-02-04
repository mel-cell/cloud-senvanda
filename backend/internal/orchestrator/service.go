package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"

	"github.com/senvanda/backend/internal/infrastructure/caddy"
	"github.com/senvanda/backend/internal/infrastructure/docker"
	"github.com/senvanda/backend/internal/infrastructure/woodpecker"
)

type Service struct {
	app              *pocketbase.PocketBase
	dockerClient     *docker.Client
	caddyClient      *caddy.Client
	woodpeckerClient *woodpecker.Client
}

func NewService(app *pocketbase.PocketBase, dockerClient *docker.Client, caddyClient *caddy.Client, woodpeckerClient *woodpecker.Client) *Service {
	return &Service{
		app:              app,
		dockerClient:     dockerClient,
		caddyClient:      caddyClient,
		woodpeckerClient: woodpeckerClient,
	}
}

// TriggerBuildPipeline initiates a build in Woodpecker CI
func (s *Service) TriggerBuildPipeline(project *models.Record, branch string) error {
	repoOwner := project.GetString("repo_owner") // e.g. "melvin"
	repoName := project.GetString("repo_name")   // e.g. "project-senvanda"

	if repoOwner == "" || repoName == "" {
		return fmt.Errorf("project metadata missing: repo_owner or repo_name")
	}

	log.Printf("🔨 Looking up ID and Triggering Woodpecker Build for %s/%s (Branch: %s)", repoOwner, repoName, branch)

	repoID, err := s.woodpeckerClient.LookupRepoID(repoOwner, repoName)
	if err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to lookup repo ID: %v", err))
		return err
	}

	buildNum, err := s.woodpeckerClient.TriggerBuild(repoID, branch)
	if err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to trigger build: %v", err))
		return err
	}

	log.Printf("✅ Build Triggered! Build Number: #%d", buildNum)

	// Update Project Status
	project.Set("status", "building")
	project.Set("last_build_num", buildNum)
	if err := s.app.Dao().SaveRecord(project); err != nil {
		return err
	}

	return nil
}

// DeployUserApp handles the full deployment lifecycle:
// 1. Prepare (Update DB Status)
// 2. Docker Action (Pull, Remove Old, Run New)
// 3. Caddy Action (Route Traffic)
// 4. Finalize (Update DB Status)
func (s *Service) DeployUserApp(project *models.Record, imageTag string) error {
	ctx := context.Background() // Background context for the long-running process
	projectName := project.GetString("name")

	// Default to project name if no domain specified, assuming local dev environment for now
	// In production, this would be retrieved from project record
	domain := fmt.Sprintf("%s.senvanda.local", projectName)

	// Phase 1: Preparation
	log.Printf("🚀 Starting deployment for %s...", projectName)
	project.Set("status", "deploying")
	project.Set("current_action", "🚀 Initializing deployment...")
	s.app.Dao().SaveRecord(project)

	// Phase 2: Docker Action
	containerName := fmt.Sprintf("senvanda-app-%s", projectName)
	networkName := "senvanda-apps" // Dedicated isolated network

	// A. Pull Image
	log.Printf("📦 Pulling image: %s", imageTag)
	project.Set("current_action", "📦 Pulling latest docker image...")
	s.app.Dao().SaveRecord(project)
	
	if err := s.dockerClient.PullImage(ctx, imageTag); err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to pull image: %v", err))
		return err
	}

	// B. Remove Old Container
	log.Printf("♻️ Removing old container: %s", containerName)
	project.Set("current_action", "♻️ Rotating containers...")
	s.app.Dao().SaveRecord(project)
	_ = s.dockerClient.RemoveContainer(ctx, containerName) // Ignore error if not exists

	// C. Run New Container
	log.Printf("▶️ Starting new container...")
	project.Set("current_action", "▶️ Starting new container...")
	s.app.Dao().SaveRecord(project)

	// Extract Volumes from DB
	var binds []string
	if volumesData := project.Get("volumes"); volumesData != nil {
		if volList, ok := volumesData.([]interface{}); ok {
			for _, v := range volList {
				if vm, ok := v.(map[string]interface{}); ok {
					host := fmt.Sprintf("%v", vm["host"])
					container := fmt.Sprintf("%v", vm["container"])
					if host != "" && container != "" {
						binds = append(binds, fmt.Sprintf("%s:%s", host, container))
					}
				} else if vs, ok := v.(string); ok {
					binds = append(binds, vs)
				}
			}
		}
	}

	// Extract Resources from Settings
	var cpu float64 = 0.5  // Default
	var memory int64 = 512 // Default (MB)

	if settingsData := project.Get("settings"); settingsData != nil {
		if settings, ok := settingsData.(map[string]interface{}); ok {
			if resources, ok := settings["resources"].(map[string]interface{}); ok {
				// Parse CPU
				if cpuVal, ok := resources["cpu"].(string); ok {
					fmt.Sscanf(cpuVal, "%f", &cpu)
				} else if cpuVal, ok := resources["cpu"].(float64); ok {
					cpu = cpuVal
				}

				// Parse Memory (expecting "512MB" or 512)
				if memVal, ok := resources["memory"].(string); ok {
					fmt.Sscanf(memVal, "%d", &memory)
				} else if memVal, ok := resources["memory"].(float64); ok {
					memory = int64(memVal)
				}
			}
		}
	}

	containerIP, err := s.dockerClient.RunContainer(ctx, containerName, imageTag, networkName, binds, cpu, memory)
	if err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to start container: %v", err))
		return err
	}
	log.Printf("✅ Container started at %s", containerIP)

	// Phase 3: Caddy Action
	// Retrieve port from DB, default to 80
	appPort := project.GetInt("port")
	if appPort == 0 {
		appPort = 80
	}
	target := fmt.Sprintf("%s:%d", containerIP, appPort)

	log.Printf("📡 Configuring Caddy route for %s -> %s", domain, target)
	project.Set("current_action", "📡 Configuring secure proxy...")
	s.app.Dao().SaveRecord(project)

	if err := s.caddyClient.AddLinkDomain(domain, target); err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to configure Caddy: %v", err))
		return err
	}

	// Phase 4: Finalize
	project.Set("status", "online")
	project.Set("last_deployed", time.Now())
	project.Set("internal_ip", containerIP)
	project.Set("url", fmt.Sprintf("http://%s", domain))
	project.Set("current_action", "") // Clear action on success

	if err := s.app.Dao().SaveRecord(project); err != nil {
		log.Printf("⚠️ Failed to save final project state: %v", err)
	}

	log.Printf("🎉 Deployment for %s completed successfully!", projectName)
	return nil
}

func (s *Service) markFailed(project *models.Record, reason string) {
	log.Printf("❌ Deployment failed: %s", reason)
	project.Set("status", "failed")
	project.Set("error_log", reason)
	project.Set("current_action", "❌ Deployment failed")
	s.app.Dao().SaveRecord(project)
}
