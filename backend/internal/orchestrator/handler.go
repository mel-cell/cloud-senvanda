package orchestrator

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

type DeploymentHandler struct {
	service *Service
}

func NewDeploymentHandler(service *Service) *DeploymentHandler {
	return &DeploymentHandler{service: service}
}

type DeployFinalPayload struct {
	ProjectID string `json:"project_id"`
	ImageTag  string `json:"image_tag"`
	Status    string `json:"status"`
}

func (h *DeploymentHandler) HandleDeployFinal(c echo.Context) error {
	// 1. Security Check (The Mantra)
	secret := c.Request().Header.Get("X-Senvanda-Secret")
	if secret == "" || secret != os.Getenv("SENVANDA_SHARED_SECRET") {
		log.Printf("⛔ Unauthorized Deploy attempt from %s", c.RealIP())
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid secret"})
	}

	// 2. Parse Payload
	var payload DeployFinalPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	// 3. Fetch Project from DB
	project, err := h.service.app.Dao().FindRecordById("projects", payload.ProjectID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "project not found"})
	}

	// 4. Check Build Status
	if payload.Status != "success" {
		log.Printf("⚠️ Build failed for project %s, updating status...", payload.ProjectID)
		h.service.markFailed(project, "CI Build Reported Failure")
		return c.JSON(http.StatusOK, map[string]string{"message": "build failure recorded"})
	}

	// 4. Execute Final Deployment (Asynchronous)
	// Kita pindahkan ke goroutine agar Woodpecker tidak nunggu lama

	// Construct Full Image Name
	// Format: [registry/]owner/repo:tag
	registryHost := os.Getenv("SENVANDA_REGISTRY_HOST")
	baseImage := project.GetString("repo_owner") + "/" + project.GetString("repo_name") + ":" + payload.ImageTag

	fullImage := baseImage
	if registryHost != "" {
		fullImage = registryHost + "/" + baseImage
	}

	go func() {
		if err := h.service.DeployUserApp(project, fullImage); err != nil {
			log.Printf("❌ Final Deployment failure for %s: %v", project.GetString("name"), err)
		}
	}()

	return c.JSON(http.StatusAccepted, map[string]string{
		"message": "Victory! Deployment process started.",
	})
}

func (h *DeploymentHandler) HandleAction(c echo.Context) error {
	// 1. Security Check (Secret OR PocketBase Auth)
	secret := c.Request().Header.Get("X-Senvanda-Secret")
	isAuthorized := false

	if secret != "" && secret == os.Getenv("SENVANDA_SHARED_SECRET") {
		isAuthorized = true
	} else {
		// Check PocketBase Auth
		authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
		if authRecord != nil || admin != nil {
			isAuthorized = true
		}
	}

	if !isAuthorized {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id := c.PathParam("id")
	action := c.QueryParam("action")
	force := c.QueryParam("force") == "true"

	project, err := h.service.app.Dao().FindRecordById("projects", id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "project not found"})
	}

	log.Printf("🕹️ Action received: %s for %s", action, project.GetString("name"))

	switch action {
	case "stop":
		if err := h.service.StopProject(project); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	case "delete":
		if err := h.service.DeleteProject(project); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	case "redeploy":
		if force {
			// Trigger build again from source
			branch := project.GetString("settings.branch")
			if branch == "" {
				branch = "main"
			}
			if err := h.service.TriggerBuildPipeline(project, branch); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		} else {
			// Just re-run the current image
			image := project.GetString("image")
			if image == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "no image found to redeploy"})
			}
			go h.service.DeployUserApp(project, image)
		}
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid action"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Action executed successfully"})
}

func (h *DeploymentHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/deploy-final", h.HandleDeployFinal)
	g.POST("/project/:id/action", h.HandleAction)
}
