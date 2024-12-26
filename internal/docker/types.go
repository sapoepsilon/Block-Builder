package docker

// Container represents a Docker container
type Container struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Image       string            `json:"image"`
    Status      string            `json:"status"`
    State       string            `json:"state"`
    Created     int64             `json:"created"`
    Ports       []Port            `json:"ports"`
    Labels      map[string]string `json:"labels"`
    NetworkMode string            `json:"networkMode"`
}

// Port represents a container port mapping
type Port struct {
    PrivatePort uint16 `json:"privatePort"`
    PublicPort  uint16 `json:"publicPort"`
    Type        string `json:"type"`
}
