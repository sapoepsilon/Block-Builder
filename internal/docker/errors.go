package docker

import (
	"errors"
	"strings"
)

var (
	// ErrContainerNotFound is returned when a container is not found
	ErrContainerNotFound = errors.New("container not found")

	// ErrImageNotFound is returned when an image is not found
	ErrImageNotFound = errors.New("image not found")

	// ErrContainerAlreadyExists is returned when attempting to create a container with a name that already exists
	ErrContainerAlreadyExists = errors.New("container already exists")

	// ErrInvalidConfig is returned when container configuration is invalid
	ErrInvalidConfig = errors.New("invalid container configuration")
)

// IsContainerNotFoundError checks if the error is a container not found error
func IsContainerNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container")
}

// IsImageNotFoundError checks if the error is an image not found error
func IsImageNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such image")
}

// IsResourceConstraintError checks if the error is related to resource constraints
func IsResourceConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Resource constraints exceeded")
}

// ParseContainerError parses Docker API errors and returns appropriate error types
func ParseContainerError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case IsContainerNotFoundError(err):
		return ErrContainerNotFound
	case IsImageNotFoundError(err):
		return ErrImageNotFound
	case strings.Contains(err.Error(), "Conflict"):
		return ErrContainerAlreadyExists
	default:
		return err
	}
}

// ValidateContainerConfig validates container configuration
func ValidateContainerConfig(config ContainerConfig) error {
	if config.Image == "" {
		return errors.New("image name is required")
	}

	if config.MemoryLimit < 0 {
		return errors.New("memory limit must be non-negative")
	}

	if config.CPUShares < 0 {
		return errors.New("CPU shares must be non-negative")
	}

	if config.NetworkMode != "" {
		validModes := map[string]bool{
			"bridge":     true,
			"host":       true,
			"none":       true,
			"container":  true,
		}

		// Check if it's a container network mode (container:<name|id>)
		if strings.HasPrefix(config.NetworkMode, "container:") {
			return nil
		}

		if !validModes[config.NetworkMode] {
			return errors.New("invalid network mode")
		}
	}

	if config.RestartPolicy != "" {
		validPolicies := map[string]bool{
			"no":              true,
			"always":          true,
			"unless-stopped":  true,
			"on-failure":      true,
		}

		if !validPolicies[config.RestartPolicy] {
			return errors.New("invalid restart policy")
		}
	}

	return nil
}
