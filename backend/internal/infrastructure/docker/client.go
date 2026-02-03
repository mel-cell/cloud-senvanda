package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// Client wraps the official Docker client
type Client struct {
	cli *client.Client
}

// NewClient initializes a new Docker client connecting via local socket
func NewClient() (*Client, error) {
	// client.WithAPIVersionNegotiation() otomatis memilih versi API yang cocok
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

func (c *Client) GetRawClient() *client.Client {
	return c.cli
}

// Ping checks connectivity to the Docker Daemon
func (c *Client) Ping(ctx context.Context) (types.Ping, error) {
	return c.cli.Ping(ctx)
}

// ListContainers returns a list of all containers (running and stopped)
func (c *Client) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{All: true})
}

// PullImage pulls the latest version of the image
func (c *Client) PullImage(ctx context.Context, imageStr string) error {
	reader, err := c.cli.ImagePull(ctx, imageStr, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	// Discard output to avoid filling memory/logs, or we could pipe it to a logger
	// For now, we just wait for it to finish
	// io.Copy(io.Discard, reader) -> but we need 'io' package
	// Simple implementation: read until EOF
	buf := make([]byte, 1024)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			break
		}
	}
	return nil
}

// RemoveContainer stops and removes a container by name
func (c *Client) RemoveContainer(ctx context.Context, containerName string) error {
	// First convert name to ID or just use name (Docker API supports both usually)
	// But to be safe, we just try to remove.
	// Force=true means kill if running
	return c.cli.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})
}

// RunContainer creates and starts a new container attached to a specific network
// Returns the internal IP address of the container
func (c *Client) RunContainer(ctx context.Context, containerName string, image string, networkName string, binds []string, cpu float64, memory int64) (string, error) {
	// 1. Create Container
	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	// Apply Resource Limits (Stability Pillar)
	if cpu > 0 {
		hostConfig.Resources.NanoCPUs = int64(cpu * 1e9)
	}
	if memory > 0 {
		hostConfig.Resources.Memory = memory * 1024 * 1024 // Convert MB to Bytes
	}

	resp, err := c.cli.ContainerCreate(ctx,
		&container.Config{
			Image:    image,
			Hostname: containerName,
		},
		hostConfig,
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkName: {},
			},
		},
		nil,
		containerName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 2. Start Container
	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// 3. Inspect to get IP
	inspect, err := c.cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get IP from the specified network
	netSettings, ok := inspect.NetworkSettings.Networks[networkName]
	if !ok {
		return "", fmt.Errorf("container started but not connected to network %s", networkName)
	}

	return netSettings.IPAddress, nil
}

// Close closes the transport
func (c *Client) Close() error {
	return c.cli.Close()
}
