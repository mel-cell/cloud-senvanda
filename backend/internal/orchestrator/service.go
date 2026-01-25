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
	if err := s.app.Dao().SaveRecord(project); err != nil {
		log.Printf("⚠️ Failed to update project status: %v", err)
		// Continue anyway? For now, yes.
	}

	// Phase 2: Docker Action
	containerName := fmt.Sprintf("senvanda-app-%s", projectName)
	networkName := "senvanda-apps" // Dedicated isolated network

	// A. Pull Image
	log.Printf("📦 Pulling image: %s", imageTag)
	if err := s.dockerClient.PullImage(ctx, imageTag); err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to pull image: %v", err))
		return err
	}

	// B. Remove Old Container
	log.Printf("♻️ Removing old container: %s", containerName)
	_ = s.dockerClient.RemoveContainer(ctx, containerName) // Ignore error if not exists

	// C. Run New Container
	log.Printf("▶️ Starting new container...")
	containerIP, err := s.dockerClient.RunContainer(ctx, containerName, imageTag, networkName)
	if err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to start container: %v", err))
		return err
	}
	log.Printf("✅ Container started at %s", containerIP)

	// Phase 3: Caddy Action
	// Assume user app runs on port 80 inside the container
	target := fmt.Sprintf("%s:80", containerIP)

	log.Printf("📡 Configuring Caddy route for %s -> %s", domain, target)
	if err := s.caddyClient.AddLinkDomain(domain, target); err != nil {
		s.markFailed(project, fmt.Sprintf("Failed to configure Caddy: %v", err))
		return err
	}

	// Phase 4: Finalize
	project.Set("status", "online")
	project.Set("last_deployed", time.Now())
	project.Set("internal_ip", containerIP)
	project.Set("url", fmt.Sprintf("http://%s", domain))

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
	s.app.Dao().SaveRecord(project)
}
