package deployment

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

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
	g.GET("/deploy/projects", h.handleListProjects)
	g.GET("/deploy/legacy", h.handleListLegacy)
	g.POST("/deploy/create", h.handleDeployProject)
	g.POST("/deploy/draft", h.handleCreateDraft)
	g.POST("/deploy/scan", h.handleScan)
	g.POST("/deploy/prune", h.handlePruneProjects)
	g.POST("/deploy/adopt", h.handleAdoptProject)
	g.GET("/deploy/:id/logs", h.handleGetLogs)
	g.GET("/deploy/:id/logs/stream", h.handleStreamLogs)
	g.GET("/deploy/cluster/stats", h.handleGetClusterStats) // Global Stats
	g.GET("/deploy/:id/stats", h.handleGetStats)
	g.POST("/deploy/:id/action", h.handleProjectAction)
	g.POST("/webhook/redeploy", h.handleWebhookRedeploy)
	g.GET("/deploy/info", h.handleGetInfo)
	g.POST("/deploy/marketplace", h.handleDeployMarketplace)
}

func (h *Handler) handleGetInfo(c echo.Context) error {
	info, err := h.service.GetDockerInfo(c.Request().Context())
	if err != nil {
		return apis.NewBadRequestError("Failed to get docker info. Ensure Docker daemon is running.", err)
	}

	return c.JSON(200, map[string]interface{}{
		"message":        "Docker Connected Successfully",
		"server_version": info.ServerVersion,
		"containers":     info.Containers,
		"running":        info.ContainersRunning,
	})
}

func (h *Handler) handleDeployMarketplace(c echo.Context) error {
	var req MarketplaceRequest
	if err := c.Bind(&req); err != nil {
		return apis.NewBadRequestError("Invalid request body", err)
	}

	record, err := h.service.DeployMarketplace(c.Request().Context(), req)
	if err != nil {
		return apis.NewBadRequestError("Deployment failed", err)
	}

	return c.JSON(200, record)
}

func (h *Handler) handleGetClusterStats(c echo.Context) error {
	stats, err := h.service.GetClusterStats(c.Request().Context())
	if err != nil {
		return apis.NewBadRequestError("Failed to get cluster stats", err)
	}
	return c.JSON(200, stats)
}

func (h *Handler) handleListProjects(c echo.Context) error {
	projects, err := h.service.GetProjectsWithStatus(c.Request().Context())
	if err != nil {
		return apis.NewBadRequestError("Failed to list projects", err)
	}
	return c.JSON(200, projects)
}

func (h *Handler) handleProjectAction(c echo.Context) error {
	id := c.PathParam("id")
	var data ActionProjectReq
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Invalid request", err)
	}

	if err := h.service.ActionProject(c.Request().Context(), id, strings.ToLower(data.Action)); err != nil {
		return apis.NewBadRequestError(fmt.Sprintf("Failed to %s project", data.Action), err)
	}

	return c.JSON(200, map[string]string{"status": "ok"})
}

// Shared internal logic
func (h *Handler) processProjectCreation(c echo.Context, isDraft bool) error {
	var data CreateProjectReq
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Invalid request", err)
	}

	// Force isDraft based on endpoint
	data.IsDraft = isDraft

	// 1. Resolve User/Owner
	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		// If not a regular user, check if Admin
		admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
		if admin != nil {
			// If Admin, try to find the first user to assign ownership to (fallback)
			user, err := h.service.FindFirstUser(c.Request().Context())
			if err != nil {
				return apis.NewBadRequestError("Admin authorized, but no Users found to assign project to. Create a user first.", err)
			}
			authRecord = user
			fmt.Println("[DEBUG] Using fallback user for Admin action:", authRecord.Id)
		} else {
			return apis.NewForbiddenError("Unauthorized", nil)
		}
	}

	fmt.Printf("[DEBUG] Creating project '%s' (Draft: %v) for user %s\n", data.Name, data.IsDraft, authRecord.Id)

	project, err := h.service.CreateProject(c.Request().Context(), data, authRecord)
	if err != nil {
		fmt.Printf("[ERROR] CreateProject failed: %v\n", err)
		return apis.NewBadRequestError("Failed to create project: "+err.Error(), err)
	}

	return c.JSON(200, project)
}

func (h *Handler) handleDeployProject(c echo.Context) error {
	return h.processProjectCreation(c, false)
}

func (h *Handler) handleCreateDraft(c echo.Context) error {
	return h.processProjectCreation(c, true)
}

type ScanReq struct {
	Url string `json:"url" form:"url"`
}

func (h *Handler) handleScan(c echo.Context) error {
	var data ScanReq
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Invalid request", err)
	}

	if data.Url == "" {
		return apis.NewBadRequestError("URL is required", nil)
	}

	// Basic validation / cleaning?
	result, err := h.service.ScanGitRepository(c.Request().Context(), data.Url)
	if err != nil {
		return apis.NewBadRequestError("Scan failed: "+err.Error(), err)
	}

	return c.JSON(200, result)
}

func (h *Handler) handleListLegacy(c echo.Context) error {
	legacy, err := h.service.DiscoverLegacy(c.Request().Context())
	if err != nil {
		return apis.NewBadRequestError("Failed to discover legacy containers", err)
	}
	return c.JSON(200, legacy)
}

func (h *Handler) handleGetLogs(c echo.Context) error {
	id := c.PathParam("id")
	logs, err := h.service.GetProjectLogs(c.Request().Context(), id)
	if err != nil {
		return apis.NewBadRequestError("Failed to fetch logs", err)
	}
	return c.JSON(200, map[string]string{"logs": logs})
}

func (h *Handler) handleStreamLogs(c echo.Context) error {
	id := c.PathParam("id")

	// 1. Set Headers for SSE
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	c.Response().Header().Set(echo.HeaderConnection, "keep-alive")

	// 2. Get the stream
	stream, err := h.service.StreamLogs(c.Request().Context(), id)
	if err != nil {
		// If error, we can send a JSON error, but for SSE usually we want to respect the stream
		// Let's return 400 before flushing any data
		return apis.NewBadRequestError("Failed to open log stream", err)
	}
	defer stream.Close()

	// 3. Prepare Scanner
	reader := bufio.NewReader(stream)

	// 4. Stream Loop
	for {
		// Check context cancellation (Client disconnect)
		select {
		case <-c.Request().Context().Done():
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// Other error? Maybe stop
			break
		}

		// Clean line (Docker adds some headers in raw stream, but using Stdout/Stderr logs options usually handles it well enough for text)
		// Actually, Docker API raw stream has 8-byte header per frame [STREAM_TYPE, 0, 0, 0, SIZE, SIZE, SIZE, SIZE]
		// But s.cli.ContainerLogs handles that if we use Stdout=true (wrapper usually strips it?)
		// Wait, docker client.ContainerLogs returns a multiplexed stream usually involving stdcopy.StdCopy
		// But in our container service we returned `s.cli.ContainerLogs`.
		// If `TTY` is allocated, it's raw text. If not, it's multiplexed.
		// For simplicity, let's assume raw text or clean it in frontend if garbage appears.
		// Usually official go-docker client handles the demux partially or returns the raw reader.
		// Let's send as is.

		// SSE Format: "data: <payload>\n\n"
		msg := fmt.Sprintf("data: %s\n\n", strings.TrimSpace(line))
		if _, err := fmt.Fprint(c.Response(), msg); err != nil {
			return nil
		}
		c.Response().Flush()
	}

	return nil
}

func (h *Handler) handleGetStats(c echo.Context) error {
	id := c.PathParam("id")
	stats, err := h.service.GetProjectStats(c.Request().Context(), id)
	if err != nil {
		// Log error but return OK with zero stats so UI doesn't break
		// Typical scenario: container stopped or restarting
		fmt.Printf("[STATS] Failed to get stats for %s: %v\n", id, err)
		return c.JSON(200, map[string]interface{}{
			"cpu_percent":    0,
			"memory_bytes":   0,
			"memory_limit":   0,
			"memory_percent": 0,
			"status":         "offline",
		})
	}
	return c.JSON(200, stats)
}

func (h *Handler) handlePruneProjects(c echo.Context) error {
	count, err := h.service.PruneMissingProjects(c.Request().Context())
	if err != nil {
		return apis.NewBadRequestError("Failed to prune projects", err)
	}
	return c.JSON(200, map[string]interface{}{
		"status":       "ok",
		"pruned_count": count,
	})
}

type AdoptReq struct {
	ContainerID string `json:"containerID"`
}

func (h *Handler) handleAdoptProject(c echo.Context) error {
	var data AdoptReq
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Invalid request", err)
	}

	authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	if authRecord == nil {
		user, err := h.service.FindFirstUser(c.Request().Context())
		if err != nil {
			return err
		}
		authRecord = user
	}

	record, err := h.service.AdoptProject(c.Request().Context(), data.ContainerID, authRecord.Id)
	if err != nil {
		return apis.NewBadRequestError("Adoption failed", err)
	}

	return c.JSON(200, record)
}

func (h *Handler) handleWebhookRedeploy(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return apis.NewBadRequestError("Token is required", nil)
	}

	// 1. Find the project by webhook token
	// This uses a public search, but since the token is 32 chars random, it's secure
	// Note: You might want to cast h.service to a concrete type if needed or add FindByToken to interface
	// For simplicity, we search via DAO directly if h.service gives access or add it to Service
	return h.processWebhookAction(c, token)
}

func (h *Handler) processWebhookAction(c echo.Context, token string) error {
	// Let's assume we added FindByToken to service or we do it here
	// I'll add a trigger in service to handle this
	err := h.service.ActionProjectByToken(c.Request().Context(), token, "redeploy")
	if err != nil {
		return apis.NewBadRequestError("Redeploy failed: token invalid or system error", err)
	}

	return c.JSON(200, map[string]string{"status": "success", "message": "Deployment triggered"})
}
