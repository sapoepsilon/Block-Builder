package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

// Client wraps the Docker client
type Client struct {
	cli *client.Client
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

// ContainerConfig represents the configuration for creating a container
type ContainerConfig struct {
	Image         string
	Command       []string
	Env           []string
	WorkingDir    string
	CPUShares     int64
	MemoryLimit   int64
	NetworkMode   string
	RestartPolicy string
	Labels        map[string]string
	Ports         map[string]string // Format: "containerPort:hostPort", e.g., "3000:3000"
}

// ContainerInfo represents container information
type ContainerInfo struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	ImageID         string            `json:"image_id"`
	Command         string            `json:"command"`
	State           string            `json:"state"`
	Status          string            `json:"status"`
	Created         time.Time         `json:"created"`
	Started         time.Time         `json:"started"`
	Finished        time.Time         `json:"finished"`
	Ports           []types.Port      `json:"ports"`
	Labels          map[string]string `json:"labels"`
	SizeRw          int64            `json:"size_rw"`
	SizeRootFs      int64            `json:"size_root_fs"`
	RestartCount    int              `json:"restart_count"`
	Platform        string           `json:"platform"`
	NetworkSettings NetworkInfo       `json:"network_settings"`
	Mounts          []Mount           `json:"mounts"`
	HostConfig      HostConfig        `json:"host_config"`
	ExitCode        int               `json:"exit_code"`
}

// NetworkInfo represents container network settings
type NetworkInfo struct {
	Networks    map[string]EndpointSettings `json:"networks"`
	IPAddress   string                      `json:"ip_address"`
	Gateway     string                      `json:"gateway"`
	MacAddress  string                      `json:"mac_address"`
}

// EndpointSettings represents network endpoint settings
type EndpointSettings struct {
	IPAddress   string   `json:"ip_address"`
	Gateway     string   `json:"gateway"`
	MacAddress  string   `json:"mac_address"`
	NetworkID   string   `json:"network_id"`
	Aliases     []string `json:"aliases"`
}

// Mount represents a container mount point
type Mount struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
}

// HostConfig represents container host configuration
type HostConfig struct {
	NetworkMode    string `json:"network_mode"`
	RestartPolicy  struct {
		Name              string `json:"name"`
		MaximumRetryCount int    `json:"maximum_retry_count"`
	} `json:"restart_policy"`
	AutoRemove bool  `json:"auto_remove"`
	Memory     int64 `json:"memory"`
	CPUShares  int64 `json:"cpu_shares"`
	CPUQuota   int64 `json:"cpu_quota"`
	CPUPeriod  int64 `json:"cpu_period"`
}

// CreateContainer creates a new container with the given configuration
func (c *Client) CreateContainer(ctx context.Context, name string, config ContainerConfig) (string, error) {
	// Prepare port bindings
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	// Configure default port for Node.js applications
	for containerPort, hostPort := range config.Ports {
		natPort, err := nat.NewPort("tcp", strings.Split(containerPort, "/")[0])
		if err != nil {
			return "", &ClientError{Op: "create container", Err: err, Details: "invalid port configuration"}
		}

		portBindings[natPort] = []nat.PortBinding{{
			HostIP:   "0.0.0.0",
			HostPort: hostPort,
		}}
		exposedPorts[natPort] = struct{}{}
	}

	// Create container
	cont, err := c.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        config.Image,
			Cmd:         config.Command,
			Env:         config.Env,
			WorkingDir:  config.WorkingDir,
			Labels:      config.Labels,
			ExposedPorts: exposedPorts,
		},
		&container.HostConfig{
			NetworkMode:   container.NetworkMode(config.NetworkMode),
			PortBindings: portBindings,
			Resources: container.Resources{
				Memory:    config.MemoryLimit,
				CPUShares: config.CPUShares,
			},
			RestartPolicy: container.RestartPolicy{
				Name: container.RestartPolicyMode(config.RestartPolicy),
			},
		},
		nil,
		nil,
		name,
	)

	if err != nil {
		return "", &ClientError{
			Op:      "create_container",
			Err:     err,
			Details: "failed to create container",
		}
	}

	for _, warning := range cont.Warnings {
		fmt.Printf("Warning during container creation: %s\n", warning)
	}

	return cont.ID, nil
}

// StartContainer starts a container
func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

// ListContainers returns a list of containers
func (c *Client) ListContainers(ctx context.Context, all bool, labelFilter map[string]string) ([]ContainerInfo, error) {
	filterArgs := filters.NewArgs()
	for k, v := range labelFilter {
		filterArgs.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
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
			ID:      container.ID,
			Name:    container.Names[0],
			Image:   container.Image,
			Status:  container.Status,
			Created: time.Unix(container.Created, 0),
			State:   container.State,
			Labels:  container.Labels,
		})
	}

	return containerInfos, nil
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: force,
	})
}

// GetContainerLogs retrieves container logs
func (c *Client) GetContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
	options := container.LogsOptions{
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

// CopyToContainer copies files to a container
func (c *Client) CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader) error {
	return c.cli.CopyToContainer(ctx, containerID, dstPath, content, types.CopyToContainerOptions{})
}

// GetContainer returns detailed information about a specific container
func (c *Client) GetContainer(ctx context.Context, containerID string) (*ContainerInfo, error) {
	container, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		fmt.Printf("Error inspecting container %s: %v\n", containerID, err)
		if client.IsErrNotFound(err) {
			return nil, &ClientError{
				Op:      "inspect",
				Err:     err,
				Details: "Container not found",
			}
		}
		return nil, &ClientError{
			Op:  "inspect",
			Err: err,
		}
	}

	// Parse timestamps
	createdTime, _ := time.Parse(time.RFC3339Nano, container.Created)
	startedTime, _ := time.Parse(time.RFC3339Nano, container.State.StartedAt)
	finishedTime, _ := time.Parse(time.RFC3339Nano, container.State.FinishedAt)

	// Convert port bindings
	var ports []types.Port
	for privatePort, bindings := range container.NetworkSettings.Ports {
		for _, binding := range bindings {
			publicPort, _ := strconv.ParseUint(binding.HostPort, 10, 16)
			privatePortNum, _ := strconv.ParseUint(strings.Split(string(privatePort), "/")[0], 10, 16)
			ports = append(ports, types.Port{
				PrivatePort: uint16(privatePortNum),
				PublicPort:  uint16(publicPort),
				Type:        strings.Split(string(privatePort), "/")[1],
			})
		}
	}

	// Convert mounts
	var mounts []Mount
	for _, m := range container.Mounts {
		mounts = append(mounts, Mount{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			RW:          m.RW,
		})
	}

	// Convert network settings
	networks := make(map[string]EndpointSettings)
	for netName, net := range container.NetworkSettings.Networks {
		networks[netName] = EndpointSettings{
			IPAddress:   net.IPAddress,
			Gateway:     net.Gateway,
			MacAddress:  net.MacAddress,
			NetworkID:   net.NetworkID,
			Aliases:     net.Aliases,
		}
	}

	info := &ContainerInfo{
		ID:         container.ID,
		Name:       container.Name,
		Image:      container.Config.Image,
		ImageID:    container.Image,
		Command:    strings.Join(container.Config.Cmd, " "),
		Status:     container.State.Status,
		State:      container.State.Status,
		Created:    createdTime,
		Started:    startedTime,
		Finished:   finishedTime,
		Labels:     container.Config.Labels,
		Ports:      ports,
		NetworkSettings: NetworkInfo{
			Networks:    networks,
			IPAddress:   container.NetworkSettings.IPAddress,
			Gateway:     container.NetworkSettings.Gateway,
			MacAddress:  container.NetworkSettings.MacAddress,
		},
		Mounts:   mounts,
		Platform: container.Platform,
		HostConfig: HostConfig{
			NetworkMode: string(container.HostConfig.NetworkMode),
			RestartPolicy: struct {
				Name              string `json:"name"`
				MaximumRetryCount int    `json:"maximum_retry_count"`
			}{
				Name:              string(container.HostConfig.RestartPolicy.Name),
				MaximumRetryCount: container.HostConfig.RestartPolicy.MaximumRetryCount,
			},
			AutoRemove: container.HostConfig.AutoRemove,
			Memory:     container.HostConfig.Memory,
			CPUShares:  container.HostConfig.CPUShares,
			CPUQuota:   container.HostConfig.CPUQuota,
			CPUPeriod:  container.HostConfig.CPUPeriod,
		},
		RestartCount: container.RestartCount,
		ExitCode:     container.State.ExitCode,
	}

	return info, nil
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
