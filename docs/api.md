# API Documentation

## Overview
This document describes the REST API endpoints for the Block Builder service, which provides container management capabilities for Node.js projects.

## Base URL
```
http://localhost:8080/api/v1
```

## Endpoints

### Containers

#### Create Container
```http
POST /containers/create
```

Creates a new container for a Node.js project.

**Request Body:**
```json
{
  "projectPath": string,    // Path to Node.js project
  "name": string,          // Container name
  "env": string[],         // Environment variables (optional)
  "cpuShares": number,     // CPU shares (optional)
  "memoryLimit": number,   // Memory limit in bytes (optional)
  "networkMode": string,   // Network mode (optional)
  "labels": {             // Container labels (optional)
    "string": "string"
  }
}
```

**Response:**
- `200 OK`: Container created successfully
- `400 Bad Request`: Invalid request body or project structure
- `500 Internal Server Error`: Server error

#### List Containers
```http
GET /containers
```

Lists all containers.

**Response:**
- `200 OK`: List of containers
- `500 Internal Server Error`: Server error

#### Get Container
```http
GET /containers/{id}
```

Get container details by ID.

**Response:**
- `200 OK`: Container details
- `404 Not Found`: Container not found
- `500 Internal Server Error`: Server error

#### Get Container Logs
```http
GET /containers/{id}/logs
```

Get container logs.

**Response:**
- `200 OK`: Container logs
- `404 Not Found`: Container not found
- `500 Internal Server Error`: Server error

#### Delete Container
```http
DELETE /containers/{id}
```

Delete a container.

**Response:**
- `200 OK`: Container deleted
- `404 Not Found`: Container not found
- `500 Internal Server Error`: Server error

## Error Responses
All error responses follow this format:
```json
{
  "error": string,    // Error message
  "details": string   // Additional error details (optional)
}
```

## Rate Limiting
API requests are limited to 100 requests per minute per IP address.
