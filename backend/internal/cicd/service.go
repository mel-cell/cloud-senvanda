package cicd

import (
	"crypto/rand"
	"encoding/hex"
)

type Service interface {
	GenerateWebhookToken() string
	ValidateWebhookToken(storedToken, providedToken string) bool
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) GenerateWebhookToken() string {
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	return hex.EncodeToString(tokenBytes)
}

func (s *service) ValidateWebhookToken(storedToken, providedToken string) bool {
	if storedToken == "" || providedToken == "" {
		return false
	}
	return storedToken == providedToken
}
