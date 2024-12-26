package nodeproject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProject(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a valid package.json
	validPkgJSON := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"express": "^4.17.1"
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(validPkgJSON), 0644); err != nil {
		t.Fatalf("Failed to create test package.json: %v", err)
	}

	tests := []struct {
		name      string
		setupFunc func() string
		wantErr   bool
	}{
		{
			name: "valid project",
			setupFunc: func() string {
				return tmpDir
			},
			wantErr: false,
		},
		{
			name: "missing package.json",
			setupFunc: func() string {
				emptyDir := filepath.Join(tmpDir, "empty")
				os.MkdirAll(emptyDir, 0755)
				return emptyDir
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectPath := tt.setupFunc()
			handler := NewProjectHandler(projectPath, nil)
			err := handler.ValidateProject()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateProjectStructure(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewProjectHandler(tmpDir, nil)

	if err := handler.CreateProjectStructure(); err != nil {
		t.Fatalf("CreateProjectStructure failed: %v", err)
	}

	// Check if package.json was created
	if _, err := os.Stat(filepath.Join(tmpDir, "package.json")); err != nil {
		t.Errorf("package.json not created: %v", err)
	}

	// Check if Dockerfile was created
	if _, err := os.Stat(filepath.Join(tmpDir, "Dockerfile")); err != nil {
		t.Errorf("Dockerfile not created: %v", err)
	}
}

func TestGenerateDockerfile(t *testing.T) {
	tmpDir := t.TempDir()
	handler := NewProjectHandler(tmpDir, &ProjectConfig{
		BaseImage: "node:18-alpine",
		DefaultPort: "3000",
	})

	if err := handler.GenerateDockerfile(); err != nil {
		t.Fatalf("GenerateDockerfile failed: %v", err)
	}

	// Read generated Dockerfile
	dockerfileContent, err := os.ReadFile(filepath.Join(tmpDir, "Dockerfile"))
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}

	// Check if Dockerfile contains expected content
	expectedContent := []string{
		"FROM node:18-alpine",
		"WORKDIR /app",
		"EXPOSE 3000",
	}

	content := string(dockerfileContent)
	for _, expected := range expectedContent {
		if !contains(content, expected) {
			t.Errorf("Dockerfile missing expected content: %s", expected)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
    return strings.Contains(s, substr)
}
