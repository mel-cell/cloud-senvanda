package main

import (
    "log"

    "github.com/docker/docker/client"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
    
    "github.com/senvanda/backend/internal/deployment"
)

func main() {
    app := pocketbase.New()

    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        // 1. Init Docker Client
        dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
        if err != nil {
            return err
        }
        
        // 2. Init Modules
        deployService := deployment.NewService(dockerCli)
        deployHandler := deployment.NewHandler(deployService)

        // 3. Register Routes
        // Buat group khusus /api/senvanda
        // Pasang Middleware: RequireAdminOrRecordAuth()
        // Artinya hanya user login (Admin atau Auth Collection) yang bisa akses
        g := e.Router.Group("/api/senvanda", apis.RequireAdminOrRecordAuth())
        
        deployHandler.RegisterRoutes(g)

        return nil
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
