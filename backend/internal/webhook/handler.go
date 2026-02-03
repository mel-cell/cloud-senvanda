package webhook

import (
	"log"
	"net/http"
	"strings"

	"strings"

	"github.com/labstack/echo/v5"
	"github.com/senvanda/backend/internal/orchestrator"
)

type Handler struct {
	service      *Service
	orchestrator *orchestrator.Service
}

func NewHandler(service *Service, orchestrator *orchestrator.Service) *Handler {
	return &Handler{
		service:      service,
		orchestrator: orchestrator,
	}
}

// HandleGiteaPush receives POST requests from Gitea
// URL: POST /api/senvanda/webhook/:token
func (h *Handler) HandleGiteaPush(c echo.Context) error {
	token := c.PathParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
	}

	// 1. Parse Payload
	var payload GiteaPushPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload format"})
	}

	log.Printf("📥 Webhook Received for repo: %s (Ref: %s)", payload.Repository.FullName, payload.Ref)

	// 2. Validate Token & Logic
	project, err := h.service.ValidateTrigger(token, payload)
	if err != nil {
		log.Printf("⛔ Webhook Rejected: %v", err)
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	log.Printf("✅ Webhook Accepted for Project: %s (ID: %s)", project.GetString("name"), project.Id)

	// 3. Trigger Async CI/CD Pipeline
	go func() {
		log.Printf("⏳ Triggering CI Build for %s", project.GetString("name"))
		// Parse branch from Ref (e.g. "refs/heads/main" -> "main")
		branch := "main"
		if payload.Ref != "" {
			parts := strings.Split(payload.Ref, "/")
			if len(parts) >= 3 {
				branch = parts[len(parts)-1]
			}
		}

		if err := h.orchestrator.TriggerBuildPipeline(project, branch); err != nil {
			log.Printf("❌ Build Trigger failed: %v", err)
		}
	}()

	return c.JSON(http.StatusAccepted, map[string]string{
		"status":     "accepted",
		"project_id": project.Id,
		"message":    "CI Build triggered. Monitor Woodpecker for progress.",
	})
}

// RegisterRoutes registers the webhook endpoint
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/webhook/:token", h.HandleGiteaPush)
}
