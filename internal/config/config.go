package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration settings for the application
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Docker    DockerConfig    `yaml:"docker"`
	Container ContainerConfig `yaml:"container"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port            int           `yaml:"port" env:"SERVER_PORT" default:"8080"`
	ReadTimeout     time.Duration `yaml:"readTimeout" env:"SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `yaml:"writeTimeout" env:"SERVER_WRITE_TIMEOUT" default:"30s"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" env:"SERVER_SHUTDOWN_TIMEOUT" default:"10s"`
}

// DockerConfig holds Docker connection settings
type DockerConfig struct {
	Host       string `yaml:"host" env:"DOCKER_HOST" default:"unix:///var/run/docker.sock"`
	APIVersion string `yaml:"apiVersion" env:"DOCKER_API_VERSION" default:"1.41"`
	TLSVerify  bool   `yaml:"tlsVerify" env:"DOCKER_TLS_VERIFY" default:"false"`
	CertPath   string `yaml:"certPath" env:"DOCKER_CERT_PATH" default:""`
}

// ContainerConfig holds default container settings
type ContainerConfig struct {
	DefaultCPUShares     int64  `yaml:"cpuShares" env:"CONTAINER_CPU_SHARES" default:"1024"`
	DefaultMemoryLimit   int64  `yaml:"memoryLimit" env:"CONTAINER_MEMORY_LIMIT" default:"512000000"`
	DefaultNetworkMode   string `yaml:"networkMode" env:"CONTAINER_NETWORK_MODE" default:"bridge"`
	DefaultRestartPolicy string `yaml:"restartPolicy" env:"CONTAINER_RESTART_POLICY" default:"unless-stopped"`
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error for %s: %s", e.Field, e.Message)
}

// LoadConfig loads configuration from the specified YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	// If config file exists, load it
	if configPath != "" {
		if err := cfg.loadFromFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Load and override with environment variables
	if err := cfg.loadAndValidate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// NewConfig creates a new Config instance with values loaded from environment variables and defaults
func NewConfig() (*Config, error) {
	return LoadConfig("")
}

// loadFromFile loads configuration from a YAML file
func (c *Config) loadFromFile(configPath string) error {
	// Ensure the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// loadAndValidate loads configuration from environment variables and validates it
func (c *Config) loadAndValidate() error {
	// Load server config
	if err := c.loadServerConfig(); err != nil {
		return err
	}

	// Load Docker config
	if err := c.loadDockerConfig(); err != nil {
		return err
	}

	// Load container config
	if err := c.loadContainerConfig(); err != nil {
		return err
	}

	return c.validate()
}

func (c *Config) loadServerConfig() error {
	port, err := getEnvInt("SERVER_PORT", 8080)
	if err != nil {
		return &ConfigError{Field: "SERVER_PORT", Message: err.Error()}
	}
	c.Server.Port = port

	readTimeout, err := getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second)
	if err != nil {
		return &ConfigError{Field: "SERVER_READ_TIMEOUT", Message: err.Error()}
	}
	c.Server.ReadTimeout = readTimeout

	writeTimeout, err := getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return &ConfigError{Field: "SERVER_WRITE_TIMEOUT", Message: err.Error()}
	}
	c.Server.WriteTimeout = writeTimeout

	shutdownTimeout, err := getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return &ConfigError{Field: "SERVER_SHUTDOWN_TIMEOUT", Message: err.Error()}
	}
	c.Server.ShutdownTimeout = shutdownTimeout

	return nil
}

func (c *Config) loadDockerConfig() error {
	c.Docker.Host = getEnvString("DOCKER_HOST", "unix:///var/run/docker.sock")
	c.Docker.APIVersion = getEnvString("DOCKER_API_VERSION", "1.41")
	c.Docker.TLSVerify = getEnvBool("DOCKER_TLS_VERIFY", false)
	c.Docker.CertPath = getEnvString("DOCKER_CERT_PATH", "")

	return nil
}

func (c *Config) loadContainerConfig() error {
	cpuShares, err := getEnvInt64("CONTAINER_CPU_SHARES", 1024)
	if err != nil {
		return &ConfigError{Field: "CONTAINER_CPU_SHARES", Message: err.Error()}
	}
	c.Container.DefaultCPUShares = cpuShares

	memLimit, err := getEnvInt64("CONTAINER_MEMORY_LIMIT", 512000000)
	if err != nil {
		return &ConfigError{Field: "CONTAINER_MEMORY_LIMIT", Message: err.Error()}
	}
	c.Container.DefaultMemoryLimit = memLimit

	c.Container.DefaultNetworkMode = getEnvString("CONTAINER_NETWORK_MODE", "bridge")
	c.Container.DefaultRestartPolicy = getEnvString("CONTAINER_RESTART_POLICY", "unless-stopped")

	return nil
}

func (c *Config) validate() error {
	// Validate Server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return &ConfigError{Field: "Server.Port", Message: "port must be between 1 and 65535"}
	}
	if c.Server.ReadTimeout <= 0 {
		return &ConfigError{Field: "Server.ReadTimeout", Message: "must be positive"}
	}
	if c.Server.WriteTimeout <= 0 {
		return &ConfigError{Field: "Server.WriteTimeout", Message: "must be positive"}
	}

	// Validate Docker config
	if c.Docker.Host == "" {
		return &ConfigError{Field: "Docker.Host", Message: "cannot be empty"}
	}
	if c.Docker.APIVersion == "" {
		return &ConfigError{Field: "Docker.APIVersion", Message: "cannot be empty"}
	}

	// Validate Container config
	if c.Container.DefaultCPUShares < 0 {
		return &ConfigError{Field: "Container.DefaultCPUShares", Message: "must be non-negative"}
	}
	if c.Container.DefaultMemoryLimit < 0 {
		return &ConfigError{Field: "Container.DefaultMemoryLimit", Message: "must be non-negative"}
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) (int, error) {
	if value, exists := os.LookupEnv(key); exists {
		return strconv.Atoi(value)
	}
	return defaultValue, nil
}

func getEnvInt64(key string, defaultValue int64) (int64, error) {
	if value, exists := os.LookupEnv(key); exists {
		return strconv.ParseInt(value, 10, 64)
	}
	return defaultValue, nil
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		parsedValue, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return parsedValue
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	if value, exists := os.LookupEnv(key); exists {
		return time.ParseDuration(value)
	}
	return defaultValue, nil
}
