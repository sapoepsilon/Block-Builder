package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"docker-management-system/internal/docker"
	"github.com/gorilla/mux"
)

// ContainerHandler handles container-related HTTP requests
type ContainerHandler struct {
	dockerClient *docker.Client
}

// NewContainerHandler creates a new ContainerHandler instance
func NewContainerHandler(dockerClient *docker.Client) *ContainerHandler {
	return &ContainerHandler{
		dockerClient: dockerClient,
	}
}

// CreateContainerRequest represents the request body for container creation
// @Description Request body for creating a new container from a Node.js project
type CreateContainerRequest struct {
	ProjectPath    string            `json:"projectPath" example:"/path/to/nodejs/project" binding:"required" description:"Path to the Node.js project containing package.json"`
	Name          string            `json:"name" example:"my-nodejs-app" binding:"required" description:"Name for the container"`
	Env           []string          `json:"env,omitempty" example:"NODE_ENV=production,PORT=3000" description:"Environment variables for the Node.js application"`
	CPUShares     int64             `json:"cpuShares,omitempty" example:"1024" description:"CPU shares (relative weight)"`
	MemoryLimit   int64             `json:"memoryLimit,omitempty" example:"536870912" description:"Memory limit in bytes"`
	NetworkMode   string            `json:"networkMode,omitempty" example:"bridge" description:"Docker network mode"`
	Labels        map[string]string `json:"labels,omitempty" example:"environment:production" description:"Docker container labels"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// @Summary Create a new Node.js container
// @Description Creates a new container from a Node.js project. Validates project structure, generates Dockerfile, and configures the container
// @Description The project must contain a valid package.json file with name and version fields
// @Description Container will expose port 3000 by default and use 'npm start' as the entry command
// @Tags containers
// @Accept json
// @Produce json
// @Param request body CreateContainerRequest true "Node.js container configuration"
// @Success 201 {object} map[string]string "Returns container ID"
// @Failure 400 {object} ErrorResponse "Invalid request or invalid Node.js project structure"
// @Failure 500 {object} ErrorResponse "Server error or Docker operation failed"
// @Router /containers/create [post]
func (h *ContainerHandler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	var req CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate Node.js project structure
	if !isValidNodeProject(req.ProjectPath) {
		respondWithError(w, http.StatusBadRequest, "Invalid Node.js project", "Missing package.json or invalid structure")
		return
	}

	// Create Dockerfile in the project directory
	if err := createDockerfile(req.ProjectPath); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create Dockerfile", err.Error())
		return
	}

	// Read package.json to get project configuration
	packageJSON, err := os.ReadFile(filepath.Join(req.ProjectPath, "package.json"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read package.json", err.Error())
		return
	}

	var packageData map[string]interface{}
	if err := json.Unmarshal(packageJSON, &packageData); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse package.json", err.Error())
		return
	}

	// Create container configuration
	config := docker.ContainerConfig{
		Image:        "node:latest",
		Command:      []string{"npm", "start"},
		Env:          append(req.Env, fmt.Sprintf("NODE_PROJECT_NAME=%v", packageData["name"])),
		WorkingDir:   "/app",
		CPUShares:    req.CPUShares,
		MemoryLimit:  req.MemoryLimit,
		NetworkMode:  req.NetworkMode,
		Labels:       req.Labels,
		RestartPolicy: "no", // Docker restart policy: no, always, unless-stopped, on-failure
		Ports: map[string]string{
			"3000": "3000", // Map container port 3000 to host port 3000
		},
	}

	containerID, err := h.dockerClient.CreateContainer(r.Context(), req.Name, config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create container", err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]string{"containerId": containerID})
}

// @Summary List all containers
// @Description Get a list of all containers
// @Tags containers
// @Produce json
// @Success 200 {array} docker.Container
// @Failure 500 {object} ErrorResponse
// @Router /containers [get]
func (h *ContainerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.dockerClient.ListContainers(r.Context(), true, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list containers", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, containers)
}

// @Summary Get container by ID
// @Description Get detailed information about a container
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} docker.Container
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /containers/{id} [get]
func (h *ContainerHandler) GetContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	// Try to get all containers first
	containers, err := h.dockerClient.ListContainers(r.Context(), true, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list containers", err.Error())
		return
	}

	// Find container by either full ID or prefix
	var targetContainer *docker.ContainerInfo
	for _, container := range containers {
		if container.ID == containerID || strings.HasPrefix(container.ID, containerID) {
			targetContainer = &container
			break
		}
	}

	if targetContainer == nil {
		respondWithError(w, http.StatusNotFound, "Container not found", "")
		return
	}

	// Get detailed container info using the full ID
	container, err := h.dockerClient.GetContainer(r.Context(), targetContainer.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get container details", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, container)
}

// @Summary Get container logs
// @Description Get logs from a container
// @Tags containers
// @Produce plain
// @Param id path string true "Container ID"
// @Success 200 {string} string "Container logs"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /containers/{id}/logs [get]
func (h *ContainerHandler) GetContainerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	// Get the tail query parameter, default to "all"
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "all"
	}

	logs, err := h.dockerClient.GetContainerLogs(r.Context(), containerID, tail)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get container logs", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"logs": logs})
}

// @Summary Delete a container
// @Description Delete a container by ID
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /containers/{id} [delete]
func (h *ContainerHandler) DeleteContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	force := r.URL.Query().Get("force") == "true"
	
	if err := h.dockerClient.RemoveContainer(r.Context(), containerID, force); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to remove container", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func isValidNodeProject(projectPath string) bool {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		return false
	}

	// Read and parse package.json to verify it's valid
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return false
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return false
	}

	// Check for required fields
	_, hasName := packageJSON["name"]
	_, hasVersion := packageJSON["version"]
	return hasName && hasVersion
}

func createDockerfile(projectPath string) error {
	dockerfileContent := `FROM node:latest

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy project files
COPY . .

# Expose default port
EXPOSE 3000

# Start the application
CMD ["npm", "start"]
`
	return os.WriteFile(filepath.Join(projectPath, "Dockerfile"), []byte(dockerfileContent), 0644)
}

func respondWithError(w http.ResponseWriter, code int, message string, details string) {
	respondWithJSON(w, code, ErrorResponse{
		Error:   message,
		Details: details,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
