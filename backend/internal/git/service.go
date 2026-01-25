package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type EnvVar struct {
	Key   string
	Value string
}

type ScanResult struct {
	Framework     string
	Version       string
	StartCommand  string
	Port          int
	Image         string
	EnvVars       []EnvVar
	TracingLogs   []string
	SecurityHints []string
	DevOpsProfile map[string]string
}

type Service interface {
	ScanRepository(ctx context.Context, repoUrl string) (*ScanResult, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ScanRepository(ctx context.Context, repoUrl string) (*ScanResult, error) {
	tempDir := os.TempDir()
	scanID := fmt.Sprintf("scan-%d", time.Now().UnixNano())
	targetPath := filepath.Join(tempDir, scanID)

	result := &ScanResult{
		Port:          80,
		Image:         "nginx:alpine",
		TracingLogs:   []string{"Initializing Heuristic Engine...", "Cloning repository to secure buffer..."},
		DevOpsProfile: make(map[string]string),
	}

	defer os.RemoveAll(targetPath)

	start := time.Now()
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoUrl, targetPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git clone failed: %s : %s", err, string(output))
	}
	result.TracingLogs = append(result.TracingLogs, fmt.Sprintf("Clone successful in %v", time.Since(start)))

	result.TracingLogs = append(result.TracingLogs, "Analyzing project structure and heuristics...")

	// 1. DOCKERFILE HEURISTIC
	if s.exists(filepath.Join(targetPath, "Dockerfile")) {
		result.Framework = "Docker"
		result.Image = "custom-build"
		result.Port = 8080
		result.TracingLogs = append(result.TracingLogs, "Detected: Dockerfile. Switching to Native Container Build.")
		result.DevOpsProfile["Build Strategy"] = "Standard Dockerfile"
		result.DevOpsProfile["Optimization"] = "Multi-stage build suggested"
		return result, nil
	}

	// 2. NODE.JS / JAVASCRIPT HEURISTIC
	if s.exists(filepath.Join(targetPath, "package.json")) {
		result.TracingLogs = append(result.TracingLogs, "Found: package.json. Processing Node.js manifest...")
		result.Framework = "Node.js"
		result.Image = "node:18-alpine"
		result.Port = 3000
		result.StartCommand = "npm start"

		content, _ := os.ReadFile(filepath.Join(targetPath, "package.json"))
		pkgStr := string(content)

		if strings.Contains(pkgStr, "\"next\"") {
			result.Framework = "Next.js"
			result.StartCommand = "npm run start"
			result.TracingLogs = append(result.TracingLogs, "Heuristic: Next.js detected (SSR Mode)")
			result.SecurityHints = append(result.SecurityHints, "Ensure NEXTAUTH_SECRET is set for production.")
		} else if strings.Contains(pkgStr, "\"react\"") {
			result.Framework = "React (Static)"
			result.Image = "nginx:alpine"
			result.Port = 80
			result.TracingLogs = append(result.TracingLogs, "Heuristic: Single Page App (React) detected")
		} else if strings.Contains(pkgStr, "\"vue\"") {
			result.Framework = "Vue (Static)"
			result.Image = "nginx:alpine"
			result.Port = 80
			result.TracingLogs = append(result.TracingLogs, "Heuristic: Single Page App (Vue) detected")
		}

		// Look for .env.example
		if s.exists(filepath.Join(targetPath, ".env.example")) {
			result.TracingLogs = append(result.TracingLogs, "Template: found .env.example. Extracting variables...")
			envContent, _ := os.ReadFile(filepath.Join(targetPath, ".env.example"))
			for _, line := range strings.Split(string(envContent), "\n") {
				if strings.Contains(line, "=") {
					parts := strings.Split(line, "=")
					result.EnvVars = append(result.EnvVars, EnvVar{Key: parts[0], Value: ""})
				}
			}
		}

		result.DevOpsProfile["Runtime"] = "Node.js v18 LTS"
		result.DevOpsProfile["Manager"] = "NPM / Yarn detected"
	}

	// 3. GO / MOD HEURISTIC
	if s.exists(filepath.Join(targetPath, "go.mod")) {
		result.TracingLogs = append(result.TracingLogs, "Found: go.mod. Go project detected.")
		result.Framework = "Go"
		result.Image = "golang:1.21-alpine"
		result.Port = 8080
		result.DevOpsProfile["Runtime"] = "Go 1.21+"
		result.DevOpsProfile["Type"] = "Compiled Binary"
	}

	// 4. LARAVEL HEURISTIC
	if s.exists(filepath.Join(targetPath, "artisan")) {
		result.TracingLogs = append(result.TracingLogs, "Found: artisan. Laravel framework detected.")
		result.Framework = "Laravel"
		result.Image = "php:8.2-apache"
		result.Port = 80
		result.SecurityHints = append(result.SecurityHints, "Generate APP_KEY before starting container.")
		result.DevOpsProfile["Stack"] = "PHP/Apache"
	}

	result.TracingLogs = append(result.TracingLogs, "Heuristic analysis complete. DevOps Profile generated.")
	return result, nil
}

func (s *service) exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
