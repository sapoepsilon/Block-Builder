package nodeproject

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProjectHandler handles Node.js project operations
type ProjectHandler struct {
	projectPath string
	config      *ProjectConfig
}

// ProjectConfig holds Node.js project configuration
type ProjectConfig struct {
	RequiredDeps []string
	BaseImage    string
	DefaultPort  string
}

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
	Scripts      map[string]string `json:"scripts"`
}

// NewProjectHandler creates a new Node.js project handler
func NewProjectHandler(projectPath string, config *ProjectConfig) *ProjectHandler {
	if config == nil {
		config = &ProjectConfig{
			RequiredDeps: []string{"express"},
			BaseImage:    "node:18-alpine",
			DefaultPort:  "3000",
		}
	}
	return &ProjectHandler{
		projectPath: projectPath,
		config:     config,
	}
}

// ValidateProject checks if the project structure is valid
func (h *ProjectHandler) ValidateProject() error {
	// Check if package.json exists
	pkgPath := filepath.Join(h.projectPath, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		return fmt.Errorf("package.json not found: %w", err)
	}

	// Parse and validate package.json
	pkg, err := h.readPackageJSON()
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	// Validate required dependencies
	for _, dep := range h.config.RequiredDeps {
		if _, exists := pkg.Dependencies[dep]; !exists {
			return fmt.Errorf("required dependency %s not found", dep)
		}
	}

	return nil
}

// readPackageJSON reads and parses package.json
func (h *ProjectHandler) readPackageJSON() (*PackageJSON, error) {
	data, err := os.ReadFile(filepath.Join(h.projectPath, "package.json"))
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

// CreateProjectStructure creates the basic project structure
func (h *ProjectHandler) CreateProjectStructure() error {
	dirs := []string{
		"src",
		"public",
		"tests",
		"config",
	}

	for _, dir := range dirs {
		path := filepath.Join(h.projectPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GenerateDockerfile creates a Dockerfile for the project
func (h *ProjectHandler) GenerateDockerfile() error {
	dockerfile := fmt.Sprintf(`FROM %s

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE %s

CMD ["npm", "start"]
`, h.config.BaseImage, h.config.DefaultPort)

	err := os.WriteFile(filepath.Join(h.projectPath, "Dockerfile"), []byte(dockerfile), 0644)
	if err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	return nil
}

// PrepareBuildContext prepares the project for building
func (h *ProjectHandler) PrepareBuildContext() error {
	// Validate project first
	if err := h.ValidateProject(); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	// Create .dockerignore if it doesn't exist
	dockerignore := `node_modules
npm-debug.log
Dockerfile
.dockerignore
.git
.gitignore
README.md
`
	err := os.WriteFile(filepath.Join(h.projectPath, ".dockerignore"), []byte(dockerignore), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .dockerignore: %w", err)
	}

	return nil
}

// SetupEnvironment sets up the project environment
func (h *ProjectHandler) SetupEnvironment() error {
	envFile := `NODE_ENV=production
PORT=${PORT:-3000}
`
	err := os.WriteFile(filepath.Join(h.projectPath, ".env"), []byte(envFile), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}

	return nil
}
