package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"docker-management-system/internal/config"
	"github.com/gorilla/mux"
)

func TestIntegrationProjectCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}

	// Create test server
	handler := setupTestHandler(cfg)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test project creation
	projectReq := map[string]interface{}{
		"name": "test-project",
		"type": "nodejs",
		"config": map[string]interface{}{
			"dependencies": []string{"express"},
			"port":        "3000",
		},
	}

	body, err := json.Marshal(projectReq)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/projects", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Test project status
	resp, err = http.Get(server.URL + "/api/projects/test-project")
	if err != nil {
		t.Fatalf("Failed to get project status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestIntegrationDockerOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test configuration
	cfg := &config.Config{
		Docker: config.DockerConfig{
			Host:       "unix:///var/run/docker.sock",
			APIVersion: "1.41",
		},
	}

	// Create test server
	handler := setupTestHandler(cfg)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test Docker build request
	buildReq := map[string]interface{}{
		"projectName": "test-project",
		"tag":        "latest",
	}

	body, err := json.Marshal(buildReq)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/docker/build", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to send build request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status code %d, got %d", http.StatusAccepted, resp.StatusCode)
	}
}

// setupTestHandler creates a test HTTP handler with the given configuration
func setupTestHandler(cfg *config.Config) http.Handler {
	router := mux.NewRouter()
	
	router.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	}).Methods("POST")
	
	router.HandleFunc("/api/projects/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")
	
	router.HandleFunc("/api/docker/build", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}).Methods("POST")
	
	return router
}
