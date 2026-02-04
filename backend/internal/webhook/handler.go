package webhook

import (
	"log"
	"net/http"

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

	// 3. Update Status only (Wait for Native Woodpecker Webhook)
	log.Printf("⏳ Push Detected for %s. Waiting for Woodpecker to pick up...", project.GetString("name"))

	project.Set("status", "building")
	if err := h.service.app.Dao().SaveRecord(project); err != nil {
		log.Printf("⚠️ Failed to update status: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":     "accepted",
		"project_id": project.Id,
		"message":    "Push received. CI/CD pipeline should start shortly.",
	})
}

// RegisterRoutes registers the webhook endpoint
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/webhook/:token", h.HandleGiteaPush)
}
