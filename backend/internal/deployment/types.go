package deployment

import (
	"context"

	"github.com/docker/docker/api/types/system"
	"github.com/pocketbase/pocketbase/models"
)

// Service defines the interface for the high-level orchestrator
type Service interface {
	GetDockerInfo(ctx context.Context) (system.Info, error)
	CreateProject(ctx context.Context, req CreateProjectReq, user *models.Record) (*models.Record, error)
	GetProjectsWithStatus(ctx context.Context) ([]ProjectStatus, error)
	ActionProject(ctx context.Context, projectID string, action string) error
	ActionProjectByToken(ctx context.Context, token string, action string) error
	ScanGitRepository(ctx context.Context, repoUrl string) (*ScanResult, error)
	FindFirstUser(ctx context.Context) (*models.Record, error)
	GetProjectLogs(ctx context.Context, projectID string) (string, error) // NEW

	// NEW: Management & Adoption
	DiscoverLegacy(ctx context.Context) ([]LegacyApp, error)
	AdoptProject(ctx context.Context, containerID string, userID string) (*models.Record, error)
	PruneMissingProjects(ctx context.Context) (int, error)
}

type ProjectStatus struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Port     int                    `json:"port"`
	DBStatus string                 `json:"db_status"`
	Status   string                 `json:"status"`
	State    string                 `json:"state"`
	Created  interface{}            `json:"created"`
	Image    string                 `json:"image"`
	RepoUrl  string                 `json:"repoUrl"`
	Labels   map[string]interface{} `json:"labels,omitempty"`
}

type LegacyApp struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	State string `json:"state"`
	Port  []int  `json:"ports"`
}

type ScanResult struct {
	Framework     string            `json:"framework"`
	Version       string            `json:"version"`
	StartCommand  string            `json:"startCommand"`
	Port          int               `json:"port"`
	Domain        string            `json:"domain"`
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	EnvVars       []EnvVar          `json:"envVars"`
	TracingLogs   []string          `json:"tracingLogs"`
	SecurityHints []string          `json:"securityHints"`
	DevOpsProfile map[string]string `json:"devOpsProfile"`
}

type ProjectSettings struct {
	Branch       string    `json:"branch"`
	StartCommand string    `json:"startCommand"`
	EnvVars      []EnvVar  `json:"envVars"`
	Domain       string    `json:"domain"`
	Resources    Resources `json:"resources"`
}

type Resources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CreateProjectReq struct {
	Name      string          `json:"name"`
	RepoUrl   string          `json:"repoUrl"`
	Framework string          `json:"framework"`
	Image     string          `json:"image"`
	Port      int             `json:"port"`
	IsDraft   bool            `json:"isDraft"`
	Settings  ProjectSettings `json:"settings"`
}

type ActionProjectReq struct {
	Action string `json:"action"`
}
