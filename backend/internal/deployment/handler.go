package deployment

import (
    "context"

    "github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
)

// Service defines the interface for deployment operations
type Service interface {
    GetDockerInfo(ctx context.Context) (system.Info, error)
}

type service struct {
    dockerCli *client.Client
}

// NewService creates a new deployment service instance
func NewService(cli *client.Client) Service {
    return &service{
        dockerCli: cli,
    }
}

// GetDockerInfo retrieves basic information about the Docker daemon
func (s *service) GetDockerInfo(ctx context.Context) (system.Info, error) {
    return s.dockerCli.Info(ctx)
}

// Handler handles HTTP requests for deployment operations
type Handler struct {
    service Service
}

// NewHandler creates a new deployment handler
func NewHandler(s Service) *Handler {
    return &Handler{service: s}
}

// RegisterRoutes registers the deployment routes to the Echo group
func (h *Handler) RegisterRoutes(g *echo.Group) {
    g.POST("/deploy/info", h.handleGetInfo)
}

// handleGetInfo handles the GET /api/deploy/info request
func (h *Handler) handleGetInfo(c echo.Context) error {
    info, err := h.service.GetDockerInfo(c.Request().Context())
    if err != nil {
        return apis.NewBadRequestError("Failed to get docker info", err)
    }

    return c.JSON(200, map[string]interface{}{
        "message":        "Docker Connected Successfully",
        "server_version": info.ServerVersion,
        "containers":     info.Containers,
        "running":        info.ContainersRunning,
    })
}
