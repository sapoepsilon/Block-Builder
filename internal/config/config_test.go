package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := []byte(`
server:
  port: 9090
  readTimeout: 60s
docker:
  host: "tcp://localhost:2375"
container:
  cpuShares: 2048
`)
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 60*time.Second {
		t.Errorf("Expected readTimeout 60s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Docker.Host != "tcp://localhost:2375" {
		t.Errorf("Expected Docker host tcp://localhost:2375, got %s", cfg.Docker.Host)
	}
	if cfg.Container.DefaultCPUShares != 2048 {
		t.Errorf("Expected CPU shares 2048, got %d", cfg.Container.DefaultCPUShares)
	}
}

func TestNewConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("SERVER_PORT", "8081")
	os.Setenv("DOCKER_HOST", "unix:///test/docker.sock")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("DOCKER_HOST")
	}()

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	if cfg.Server.Port != 8081 {
		t.Errorf("Expected port 8081, got %d", cfg.Server.Port)
	}
	if cfg.Docker.Host != "unix:///test/docker.sock" {
		t.Errorf("Expected Docker host unix:///test/docker.sock, got %s", cfg.Docker.Host)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Docker: DockerConfig{
					Host:       "unix:///var/run/docker.sock",
					APIVersion: "1.41",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: Config{
				Server: ServerConfig{Port: -1},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
