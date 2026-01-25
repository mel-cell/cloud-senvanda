package container

import (
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Service interface {
	StartContainer(ctx context.Context, name string) error
	StopContainer(ctx context.Context, name string) error
	RestartContainer(ctx context.Context, name string) error
	RemoveContainer(ctx context.Context, name string) error
	CreateContainer(ctx context.Context, config *Config) (string, error)
	InspectContainer(ctx context.Context, name string) (types.ContainerJSON, error)
	ListContainers(ctx context.Context, all bool) ([]types.Container, error)
	DiscoverLegacyContainers(ctx context.Context) ([]LegacyContainer, error)
	GetContainerDetails(ctx context.Context, id string) (*ContainerDetails, error)
	GetContainerLogs(ctx context.Context, name string) (string, error)
	PullImage(ctx context.Context, image string) error
	BuildImage(ctx context.Context, contextPath string, tag string) error
	ContainerExists(ctx context.Context, name string) (bool, error)
}

type ContainerDetails struct {
	ID     string
	Name   string
	Image  string
	IP     string
	Env    []string
	Ports  []int
	Labels map[string]string
}

type Config struct {
	Name      string
	Image     string
	Ports     map[string]string // "80/tcp": "10001"
	Env       []string
	Labels    map[string]string
	Resources struct {
		CPU    string
		Memory string
	}
}

type LegacyContainer struct {
	ID    string
	Name  string
	Image string
	State string
	Ports []int
}

type service struct {
	cli *client.Client
}

func NewService(cli *client.Client) Service {
	return &service{cli: cli}
}

func (s *service) StartContainer(ctx context.Context, name string) error {
	return s.cli.ContainerStart(ctx, name, container.StartOptions{})
}

func (s *service) StopContainer(ctx context.Context, name string) error {
	return s.cli.ContainerStop(ctx, name, container.StopOptions{})
}

func (s *service) RestartContainer(ctx context.Context, name string) error {
	return s.cli.ContainerRestart(ctx, name, container.StopOptions{})
}

func (s *service) RemoveContainer(ctx context.Context, name string) error {
	return s.cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
}

func (s *service) CreateContainer(ctx context.Context, cfg *Config) (string, error) {
	portBindings := nat.PortMap{}
	for internal, external := range cfg.Ports {
		portBindings[nat.Port(internal)] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: external},
		}
	}

	// Ensure Image exists
	if _, _, err := s.cli.ImageInspectWithRaw(ctx, cfg.Image); err != nil {
		_ = s.PullImage(ctx, cfg.Image)
	}

	// Resources (Simple Parser)
	memoryLimit := int64(0)
	if cfg.Resources.Memory != "" {
		m := strings.ToUpper(cfg.Resources.Memory)
		if strings.HasSuffix(m, "GB") {
			val, _ := strconv.Atoi(strings.TrimSuffix(m, "GB"))
			memoryLimit = int64(val) * 1024 * 1024 * 1024
		} else if strings.HasSuffix(m, "MB") {
			val, _ := strconv.Atoi(strings.TrimSuffix(m, "MB"))
			memoryLimit = int64(val) * 1024 * 1024
		} else {
			val, _ := strconv.Atoi(m)
			memoryLimit = int64(val) * 1024 * 1024
		}
	}

	nanoCPUs := int64(0)
	if cfg.Resources.CPU != "" {
		val, _ := strconv.ParseFloat(cfg.Resources.CPU, 64)
		nanoCPUs = int64(val * 1e9)
	}

	resp, err := s.cli.ContainerCreate(ctx, &container.Config{
		Image:  cfg.Image,
		Env:    cfg.Env,
		Labels: cfg.Labels,
	}, &container.HostConfig{
		PortBindings:  portBindings,
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Resources: container.Resources{
			Memory:   memoryLimit,
			NanoCPUs: nanoCPUs,
		},
	}, nil, nil, cfg.Name)

	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (s *service) InspectContainer(ctx context.Context, name string) (types.ContainerJSON, error) {
	return s.cli.ContainerInspect(ctx, name)
}

func (s *service) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return s.cli.ContainerList(ctx, container.ListOptions{All: all})
}

func (s *service) DiscoverLegacyContainers(ctx context.Context) ([]LegacyContainer, error) {
	containers, err := s.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var legacy []LegacyContainer
	for _, c := range containers {
		isSenvanda := false
		for k := range c.Labels {
			if k == "senvanda.project" {
				isSenvanda = true
				break
			}
		}

		if !isSenvanda {
			name := ""
			if len(c.Names) > 0 {
				name = c.Names[0]
			}

			var ports []int
			for _, p := range c.Ports {
				if p.PublicPort != 0 {
					ports = append(ports, int(p.PublicPort))
				}
			}

			legacy = append(legacy, LegacyContainer{
				ID:    c.ID,
				Name:  name,
				Image: c.Image,
				State: c.State,
				Ports: ports,
			})
		}
	}
	return legacy, nil
}

func (s *service) GetContainerDetails(ctx context.Context, id string) (*ContainerDetails, error) {
	json, err := s.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	var ports []int
	for p := range json.NetworkSettings.Ports {
		ports = append(ports, p.Int())
	}

	return &ContainerDetails{
		ID:     json.ID,
		Name:   json.Name,
		Image:  json.Config.Image,
		IP:     json.NetworkSettings.IPAddress,
		Env:    json.Config.Env,
		Labels: json.Config.Labels,
		Ports:  ports,
	}, nil
}

func (s *service) GetContainerLogs(ctx context.Context, name string) (string, error) {
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "100"}
	out, err := s.cli.ContainerLogs(ctx, name, options)
	if err != nil {
		return "", err
	}
	defer out.Close()

	content, err := io.ReadAll(out)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *service) PullImage(ctx context.Context, img string) error {
	reader, err := s.cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)
	return nil
}
func (s *service) BuildImage(ctx context.Context, contextPath string, tag string) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", tag, contextPath)
	return cmd.Run()
}

func (s *service) ContainerExists(ctx context.Context, name string) (bool, error) {
	_, err := s.cli.ContainerInspect(ctx, name)
	if err == nil {
		return true, nil
	}
	if client.IsErrNotFound(err) {
		return false, nil
	}
	return false, err
}
