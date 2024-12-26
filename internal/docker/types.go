package docker

import (
	"github.com/docker/docker/api/types"
)

// Container represents a Docker container
type Container struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Status      string            `json:"status"`
	State       string            `json:"state"`
	Created     int64            `json:"created"`
	Ports       []types.Port     `json:"ports"`
	Labels      map[string]string `json:"labels"`
	NetworkMode string            `json:"networkMode"`
}
