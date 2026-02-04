package webhook

import (
	"errors"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// Service handles webhook business logic
type Service struct {
	app core.App
}

// NewService creates a new webhook service instance
func NewService(app core.App) *Service {
	return &Service{app: app}
}

// ValidateTrigger checks token validity and branch rules
func (s *Service) ValidateTrigger(token string, payload GiteaPushPayload) (*models.Record, error) {
	// 1. Find Project by Token
	project, err := s.app.Dao().FindRecordById("projects", token) // Assuming ID is token or we use FindFirstRecordByData
	if err != nil {
		// Fallback: Check if 'webhookToken' is a separate field (Best Practice)
		p, errData := s.app.Dao().FindFirstRecordByData("projects", "webhookToken", token)
		if errData != nil {
			return nil, errors.New("invalid webhook token: project not found")
		}
		project = p
	}

	// 2. Validate Branch (Defense Mechanism)
	allowedBranch := "main" // fallback
	if settingsData := project.Get("settings"); settingsData != nil {
		if settings, ok := settingsData.(map[string]interface{}); ok {
			if b, ok := settings["branch"].(string); ok && b != "" {
				allowedBranch = b
			}
		}
	}

	targetBranch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	if targetBranch != allowedBranch {
		return nil, errors.New("ignoring push: not to the configured branch (" + allowedBranch + ")")
	}

	return project, nil
}
