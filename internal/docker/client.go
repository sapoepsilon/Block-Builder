package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Client wraps the Docker client and provides high-level operations
type Client struct {
	cli *client.Client
}

// ContainerConfig represents the configuration for creating a container
type ContainerConfig struct {
	Image        string
	Command      []string
	Env          []string
	WorkingDir   string
	CPUShares    int64
	MemoryLimit  int64
	NetworkMode  string
	RestartPolicy string
	Labels       map[string]string
}

// ContainerInfo represents container information
type ContainerInfo struct {
	ID          string
	Name        string
	Image       string
	Status      string
	CreatedAt   time.Time
	State       string
	Labels      map[string]string
}

// ClientError represents Docker client operation errors
type ClientError struct {
	Op      string
	Err     error
	Details string
}

func (e *ClientError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("docker %s failed: %v (%s)", e.Op, e.Err, e.Details)
	}
	return fmt.Sprintf("docker %s failed: %v", e.Op, e.Err)
}

// NewClient creates a new Docker client
func NewClient(host, version string, tlsVerify bool, certPath string) (*Client, error) {
	opts := []client.Opt{
		client.WithHost(host),
		client.WithVersion(version),
	}

	if tlsVerify {
		opts = append(opts, client.WithTLSClientConfig(
			certPath+"/cert.pem",
			certPath+"/key.pem",
			certPath+"/ca.pem",
		))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, &ClientError{
			Op:  "connect",
			Err: err,
		}
	}

	return &Client{cli: cli}, nil
}

// CreateContainer creates a new container with the given configuration
func (c *Client) CreateContainer(ctx context.Context, name string, config ContainerConfig) (string, error) {
	containerConfig := &container.Config{
		Image:      config.Image,
		Cmd:        config.Command,
		Env:        config.Env,
		WorkingDir: config.WorkingDir,
		Labels:     config.Labels,
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			CPUShares: config.CPUShares,
			Memory:    config.MemoryLimit,
		},
		NetworkMode:   container.NetworkMode(config.NetworkMode),
		RestartPolicy: container.RestartPolicy{
			Name: config.RestartPolicy,
		},
	}

	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, name)
	if err != nil {
		return "", &ClientError{
			Op:  "create_container",
			Err: err,
		}
	}

	for _, warning := range resp.Warnings {
		// Log warnings here if you have a logger
		fmt.Printf("Warning during container creation: %s\n", warning)
	}

	return resp.ID, nil
}

// ListContainers returns a list of containers
func (c *Client) ListContainers(ctx context.Context, all bool, labelFilter map[string]string) ([]ContainerInfo, error) {
	filterArgs := filters.NewArgs()
	for k, v := range labelFilter {
		filterArgs.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     all,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, &ClientError{
			Op:  "list_containers",
			Err: err,
		}
	}

	var containerInfos []ContainerInfo
	for _, container := range containers {
		containerInfos = append(containerInfos, ContainerInfo{
			ID:        container.ID,
			Name:      container.Names[0],
			Image:     container.Image,
			Status:    container.Status,
			CreatedAt: time.Unix(container.Created, 0),
			State:     container.State,
			Labels:    container.Labels,
		})
	}

	return containerInfos, nil
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	err := c.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: force,
	})
	if err != nil {
		return &ClientError{
			Op:  "remove_container",
			Err: err,
		}
	}

	return nil
}

// GetContainerLogs retrieves container logs
func (c *Client) GetContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	logs, err := c.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", &ClientError{
			Op:  "get_logs",
			Err: err,
		}
	}
	defer logs.Close()

	// Docker multiplexes stdout and stderr, so we need to handle both streams
	var stdout, stderr io.Writer
	stdout = io.Discard
	stderr = io.Discard

	// Create buffers for stdout and stderr
	stdoutBuf := new(stdWriterBuffer)
	stderrBuf := new(stdWriterBuffer)
	stdout = stdoutBuf
	stderr = stderrBuf

	_, err = stdcopy.StdCopy(stdout, stderr, logs)
	if err != nil {
		return "", &ClientError{
			Op:  "read_logs",
			Err: err,
		}
	}

	// Combine stdout and stderr
	return fmt.Sprintf("STDOUT:\n%s\nSTDERR:\n%s", stdoutBuf.String(), stderrBuf.String()), nil
}

// Helper type for capturing container logs
type stdWriterBuffer struct {
	buffer []byte
}

func (w *stdWriterBuffer) Write(p []byte) (int, error) {
	w.buffer = append(w.buffer, p...)
	return len(p), nil
}

func (w *stdWriterBuffer) String() string {
	return string(w.buffer)
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	if err := c.cli.Close(); err != nil {
		return &ClientError{
			Op:  "close",
			Err: err,
		}
	}
	return nil
}
