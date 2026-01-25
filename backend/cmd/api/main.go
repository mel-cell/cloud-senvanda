package main

import (
	"log"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"

	"github.com/senvanda/backend/internal/cicd"
	"github.com/senvanda/backend/internal/container"
	"github.com/senvanda/backend/internal/deployment"
	"github.com/senvanda/backend/internal/git"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// 1. Init Docker Client
		dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}

		// 1b. Ensure 'projects' Collection Exists & Has Correct Schema
		colProxy, err := app.Dao().FindCollectionByNameOrId("projects")
		if err != nil {
			// Create New
			log.Println("⚡ Creating 'projects' collection...")
			colProxy = &models.Collection{}
			colProxy.Name = "projects"
			colProxy.Type = models.CollectionTypeBase
			colProxy.ListRule = nil // Allow all for now debug (or restricted)
			strPtr := func(s string) *string { return &s }
			colProxy.ListRule = strPtr("@request.auth.id != ''")
			colProxy.ViewRule = strPtr("@request.auth.id != ''")
			colProxy.CreateRule = strPtr("@request.auth.id != ''")
			colProxy.UpdateRule = strPtr("@request.auth.id != ''")
			colProxy.DeleteRule = strPtr("@request.auth.id != ''")
		} else {
			log.Println("⚡ Updating 'projects' schema...")
		}

		// Define Desired Schema & Rules (ALWAYS APPLY)
		strPtr := func(s string) *string { return &s }
		colProxy.ListRule = strPtr("@request.auth.id != ''")
		colProxy.ViewRule = strPtr("@request.auth.id != ''")
		colProxy.CreateRule = strPtr("@request.auth.id != ''")
		colProxy.UpdateRule = strPtr("@request.auth.id != ''")
		colProxy.DeleteRule = strPtr("@request.auth.id != ''")

		desiredSchema := schema.NewSchema(
			&schema.SchemaField{Name: "name", Type: schema.FieldTypeText, Required: true, Unique: true},
			&schema.SchemaField{Name: "port", Type: schema.FieldTypeNumber, Required: true},
			&schema.SchemaField{Name: "status", Type: schema.FieldTypeText},
			&schema.SchemaField{Name: "repoUrl", Type: schema.FieldTypeText},
			&schema.SchemaField{Name: "framework", Type: schema.FieldTypeText},
			&schema.SchemaField{Name: "image", Type: schema.FieldTypeText},
			&schema.SchemaField{Name: "containerId", Type: schema.FieldTypeText}, // NEW
			&schema.SchemaField{Name: "webhookToken", Type: schema.FieldTypeText, Unique: true},
			&schema.SchemaField{Name: "settings", Type: schema.FieldTypeJson},
			&schema.SchemaField{Name: "user", Type: schema.FieldTypeRelation, Options: &schema.RelationOptions{CollectionId: "users", CascadeDelete: false}},
		)
		colProxy.Schema = desiredSchema

		if err := app.Dao().SaveCollection(colProxy); err != nil {
			return err
		}

		// 2. Init Modular Services
		containerSvc := container.NewService(dockerCli)
		gitSvc := git.NewService()
		cicdSvc := cicd.NewService()

		// Allow CORS for the new frontend port
		e.Router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"http://localhost:7098", "http://127.0.0.1:7098"},
			AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		}))

		deployService := deployment.NewService(app, containerSvc, gitSvc, cicdSvc)
		deployHandler := deployment.NewHandler(deployService)

		// 3. Register Routes
		// Buat group khusus /api/senvanda
		// Pasang Middleware: RequireAdminOrRecordAuth()
		// Artinya hanya user login (Admin atau Auth Collection) yang bisa akses
		g := e.Router.Group("/api/senvanda", apis.RequireAdminOrRecordAuth())

		deployHandler.RegisterRoutes(g)

		// DEVOPS: Legacy seeding disabled to prevent zombie records.
		// Use the 'Adopt' feature from Dashboard to manage existing containers.

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Trigger Rebuild
