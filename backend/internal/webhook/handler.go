package webhook

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
		// Return 200 OK to Gitea even if rejected logic-wise to stop it from retrying,
		// but with a descriptive message. Or 403 if we want to be strict.
		// Let's go strict for now to catch config errors.
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	log.Printf("✅ Webhook Accepted for Project: %s (ID: %s)", project.GetString("name"), project.Id)

	// TODO: Phase 4 - Trigger Deployment Pipeline
	// deploymentService.Trigger(project, payload.After)

	return c.JSON(http.StatusOK, map[string]string{
		"status":     "accepted",
		"project_id": project.Id,
		"message":    "Deployment queued (Coming Soon)",
	})
}

// RegisterRoutes registers the webhook endpoint
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/webhook/:token", h.HandleGiteaPush)
}
