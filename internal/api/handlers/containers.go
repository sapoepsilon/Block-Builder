package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"github.com/gorilla/mux"
	"context"

	"github.com/your-username/block-builder/internal/docker"
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
type CreateContainerRequest struct {
	ProjectPath    string            `json:"projectPath"`
	Name          string            `json:"name"`
	Env           []string          `json:"env,omitempty"`
	CPUShares     int64             `json:"cpuShares,omitempty"`
	MemoryLimit   int64             `json:"memoryLimit,omitempty"`
	NetworkMode   string            `json:"networkMode,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// CreateContainer handles POST /containers/create
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

// ListContainers handles GET /containers
func (h *ContainerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	containers, err := h.dockerClient.ListContainers(r.Context(), true, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list containers", err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, containers)
}

// GetContainer handles GET /containers/{id}
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

// GetContainerLogs handles GET /containers/{id}/logs
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

// DeleteContainer handles DELETE /containers/{id}
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
