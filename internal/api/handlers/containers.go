package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

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
// @Description Request body for creating a new container
type CreateContainerRequest struct {
	ProjectPath    string            `json:"projectPath" example:"/path/to/nodejs/project" binding:"required"`
	Name          string            `json:"name" example:"my-app-container" binding:"required"`
	Env           []string          `json:"env,omitempty" example:"NODE_ENV=production"`
	CPUShares     int64             `json:"cpuShares,omitempty" example:"1024"`
	MemoryLimit   int64             `json:"memoryLimit,omitempty" example:"536870912"`
	NetworkMode   string            `json:"networkMode,omitempty" example:"bridge"`
	Labels        map[string]string `json:"labels,omitempty" example:"environment:production"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// @Summary Create a new container
// @Description Create a new container from a Node.js project
// @Tags containers
// @Accept json
// @Produce json
// @Param request body CreateContainerRequest true "Container configuration"
// @Success 200 {object} docker.Container
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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

	// Create container configuration
	config := docker.ContainerConfig{
		Image:        "node:latest",
		Command:      []string{"node", "index.js"},
		Env:          req.Env,
		WorkingDir:   "/app",
		CPUShares:    req.CPUShares,
		MemoryLimit:  req.MemoryLimit,
		NetworkMode:  req.NetworkMode,
		Labels:       req.Labels,
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

	containers, err := h.dockerClient.ListContainers(r.Context(), true, map[string]string{"id": containerID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get container", err.Error())
		return
	}

	if len(containers) == 0 {
		respondWithError(w, http.StatusNotFound, "Container not found", "")
		return
	}

	respondWithJSON(w, http.StatusOK, containers[0])
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
	return true
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
