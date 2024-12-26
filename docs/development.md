# Development Guide

## Setup Instructions

### Prerequisites
- Go 1.21 or later
- Docker Engine
- Node.js (for testing Node.js projects)

### Installation
1. Clone the repository:
```bash
git clone <repository-url>
cd block-builder
```

2. Install dependencies:
```bash
go mod download
```

## Configuration

### Environment Variables
The application can be configured using environment variables or a configuration file:

- `PORT`: Server port (default: 8080)
- `LOG_LEVEL`: Logging level (default: info)
- `DOCKER_HOST`: Docker daemon socket (default: unix:///var/run/docker.sock)
- `MAX_CONTAINERS`: Maximum number of containers per user (default: 10)
- `RATE_LIMIT`: API rate limit per minute (default: 100)

### Configuration File
Create a `config.yaml` in the `config` directory:
```yaml
server:
  port: 8080
  host: "0.0.0.0"

logging:
  level: "info"
  format: "json"

docker:
  host: "unix:///var/run/docker.sock"
  maxContainers: 10

security:
  rateLimit: 100
```

## Testing

### Running Tests
1. Unit tests:
```bash
go test ./...
```

2. Integration tests:
```bash
go test ./internal/api -tags=integration
```

### Test Coverage
Generate test coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Development Workflow

1. Create a new branch for your feature/fix:
```bash
git checkout -b feature/your-feature
```

2. Make your changes following the [LLM Guidelines](../llm-guidelines.md)

3. Run tests and linting:
```bash
go test ./...
golangci-lint run
```

4. Commit your changes:
```bash
git commit -m "feat: your feature description"
```

## Deployment

### Local Deployment
1. Build the binary:
```bash
go build -o block-builder ./cmd/server
```

2. Run the server:
```bash
./block-builder
```

### Docker Deployment
1. Build the Docker image:
```bash
docker build -t block-builder .
```

2. Run the container:
```bash
docker run -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock block-builder
```

## Monitoring and Logging

### Logging
- Logs are written to stdout in JSON format
- Log levels: debug, info, warn, error
- Each log entry includes:
  - Timestamp
  - Level
  - Message
  - Additional context fields

### Metrics
The application exposes Prometheus metrics at `/metrics` endpoint:
- HTTP request counters
- Response time histograms
- Container operation metrics
- System resource usage
