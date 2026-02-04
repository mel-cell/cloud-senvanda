package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/senvanda/backend/internal/cicd"
	"github.com/senvanda/backend/internal/container"
	"github.com/senvanda/backend/internal/deployment"
	"github.com/senvanda/backend/internal/git"
	"github.com/senvanda/backend/internal/infrastructure/caddy"
	"github.com/senvanda/backend/internal/infrastructure/docker"
	"github.com/senvanda/backend/internal/infrastructure/woodpecker"
	"github.com/senvanda/backend/internal/orchestrator"
	"github.com/senvanda/backend/internal/webhook"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// 0. Ensure 'projects' Collection Exists & Has Correct Schema
		col, err := app.Dao().FindCollectionByNameOrId("projects")
		if err != nil {
			log.Println("⚠️ Collection 'projects' not found, creating...")
			col = &models.Collection{}
			col.Name = "projects"
			col.Type = models.CollectionTypeBase
		}

		// Sync Schema Fields
		requiredFields := []schema.SchemaField{
			{Name: "name", Type: schema.FieldTypeText, Required: true},
			{Name: "status", Type: schema.FieldTypeText},
			{Name: "webhook_token", Type: schema.FieldTypeText},
			{Name: "repo_owner", Type: schema.FieldTypeText},
			{Name: "repo_name", Type: schema.FieldTypeText},
			{Name: "internal_ip", Type: schema.FieldTypeText},
			{Name: "last_build_num", Type: schema.FieldTypeNumber},
			{Name: "port", Type: schema.FieldTypeNumber},
			{Name: "error_log", Type: schema.FieldTypeText},
			{Name: "volumes", Type: schema.FieldTypeJson},
			{Name: "settings", Type: schema.FieldTypeJson},
			{Name: "current_action", Type: schema.FieldTypeText}, // For Real-time UX
		}

		for _, f := range requiredFields {
			if col.Schema.GetFieldByName(f.Name) == nil {
				col.Schema.AddField(&f)
			}
		}

		// PERMISSIVE RULES FOR TESTING
		rule := ""
		col.ListRule = &rule
		col.ViewRule = &rule
		col.CreateRule = &rule
		col.UpdateRule = &rule

		if err := app.Dao().SaveCollection(col); err != nil {
			log.Printf("❌ Failed to save collection: %v", err)
			return err
		}

		if err := app.Dao().SaveCollection(col); err != nil {
			return err
		}

		// SEEDING: Ensure dummy project exists for testing
		dummyProject, err := app.Dao().FindFirstRecordByData("projects", "name", "project-senvanda")
		if err != nil {
			log.Println("🌱 Seeding dummy project...")
			dummyProject = models.NewRecord(col)
			dummyProject.Set("name", "project-senvanda")
			if errSave := app.Dao().SaveRecord(dummyProject); errSave != nil {
				log.Printf("❌ Failed to seed dummy project: %v", errSave)
			} else {
				log.Printf("🔑 DUMMY TOKEN (ID): %s", dummyProject.Id)
			}
		} else {
			log.Printf("🔑 DUMMY TOKEN (ID): %s", dummyProject.Id)
		}

		// 1. Inisialisasi Infrastructure Layer (The Heart & The Receptionist)
		dockerClient, err := docker.NewClient()
		if err != nil {
			log.Printf("❌ Gagal connect ke Docker: %v", err)
			return err
		}
		defer dockerClient.Close()

		caddyURL := os.Getenv("CADDY_API_URL")
		if caddyURL == "" {
			caddyURL = "http://caddy:2019"
		}
		caddyClient := caddy.NewClient(caddyURL)

		// Woodpecker Client (The Worker)
		ciToken := os.Getenv("WOODPECKER_API_TOKEN")
		ciURL := os.Getenv("WOODPECKER_API_URL")
		if ciURL == "" {
			ciURL = "http://woodpecker-server:8000"
		}
		woodpeckerClient := woodpecker.NewClient(ciURL, ciToken)

		// 2. Inisialisasi Logic Layer (The Brain)
		orchestratorSvc := orchestrator.NewService(app, dockerClient, caddyClient, woodpeckerClient)
		deployHandler := orchestrator.NewDeploymentHandler(orchestratorSvc)

		webhookSvc := webhook.NewService(app)
		webhookHandler := webhook.NewHandler(webhookSvc, orchestratorSvc)

		// Legacy/Dashboard Support
		containerSvc := container.NewService(dockerClient.GetRawClient())
		gitSvc := git.NewService()
		cicdSvc := cicd.NewService()
		deploymentSvc := deployment.NewService(app, containerSvc, gitSvc, cicdSvc)
		deploymentHandler := deployment.NewHandler(deploymentSvc)

		// 3. Register Routes
		// Group API Public
		apiGroup := e.Router.Group("/api/senvanda")

		// Register Webhook (from Gitea)
		webhookHandler.RegisterRoutes(apiGroup)

		// Register Deploy Final (from Woodpecker)
		deployHandler.RegisterRoutes(apiGroup)

		// Register Dashboard Routes
		deploymentHandler.RegisterRoutes(apiGroup)

		// Test Endpoint (Bukti Kehidupan)
		// Bisa diakses via: GET http://localhost:8090/api/senvanda/health-check
		apiGroup.GET("/health-check", func(c echo.Context) error {
			// Caddy Ping
			caddyStatus := "online"
			if err := caddyClient.Ping(); err != nil {
				caddyStatus = fmt.Sprintf("offline (%v)", err)
			}

			// Coba Ping Docker
			ping, err := dockerClient.Ping(c.Request().Context())
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"status": "error",
					"detail": err.Error(),
				})
			}

			// Coba List Container
			containers, err := dockerClient.ListContainers(c.Request().Context())
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"status": "error",
					"detail": err.Error(),
				})
			}

			// Return Good News
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status": "online",
				"caddy":  caddyStatus,
				"docker": map[string]interface{}{
					"api_version":     ping.APIVersion,
					"container_count": len(containers),
				},
				"message": "Senvanda Backend is connected to Docker Engine & Caddy!",
			})
		})

		// TEST: Add Dummy Domain
		// POST /api/senvanda/test-caddy?domain=test.local&target=1.1.1.1:80
		apiGroup.POST("/test-caddy", func(c echo.Context) error {
			domain := c.QueryParam("domain")
			target := c.QueryParam("target")
			if domain == "" || target == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing domain or target"})
			}
			if err := caddyClient.AddLinkDomain(domain, target); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusOK, map[string]string{"message": "Route added to Caddy!"})
		})

		log.Println("✅ Senvanda v2 Control Plane is Ready!")
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
