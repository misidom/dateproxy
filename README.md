# Date-Based HTTP Proxy

A production-grade HTTP proxy server that routes requests based on date query parameters. Written in Go with Kubernetes deployment support.

## Features

- **Date-based routing**: Routes HTTP requests based on `date` query parameter in YYYYMMDD format
- **Configurable date ranges**: Define multiple date ranges and their corresponding downstream services
- **Production-ready**: Includes proper timeouts, graceful shutdown, and security best practices
- **Kubernetes support**: Complete K8s manifests with health checks, resource limits, and security contexts
- **Test services**: Includes sample backend services for testing

## Quick Start

### Local Development

1. **Build and run test services**:
```bash
make dev-setup
```

2. **In separate terminals, start the backend services**:
```bash
make run-service1  # Runs on :8081
make run-service2  # Runs on :8082
make run-service3  # Runs on :8083
```

3. **Start the proxy**:
```bash
make run  # Runs on :8080
```

4. **Test the routing**:
```bash
make test
```

### Example Requests

```bash
# Routes to service1 (2020-2022 range)
curl "http://localhost:8080/api/data?date=20210515"

# Routes to service2 (2023-2024 range)
curl "http://localhost:8080/api/data?date=20230815"

# Routes to service3 (2025-2027 range)
curl "http://localhost:8080/api/data?date=20250315"
```

## Configuration

The proxy is configured via YAML file (`config.yaml`):

```yaml
port: 8080
read_timeout: "30s"
write_timeout: "30s"
idle_timeout: "120s"

date_ranges:
  - start_date: "20200101"
    end_date: "20221231"
    service: "http://localhost:8081"
  - start_date: "20230101"
    end_date: "20241231"
    service: "http://localhost:8082"
  - start_date: "20250101"
    end_date: "20271231"
    service: "http://localhost:8083"
```

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster
- kubectl configured
- Docker images built and pushed to registry

### Deploy

```bash
# Build Docker images
make docker-build

# Deploy to Kubernetes
make k8s-deploy

# Clean up
make k8s-clean
```

### Architecture

The Kubernetes deployment includes:
- **ConfigMap**: Stores proxy configuration
- **Deployment**: 3 replicas of the proxy with health checks
- **Service**: ClusterIP service for internal access
- **Backend services**: 3 test services with 2 replicas each

## API

### Request Format

All requests must include a `date` query parameter:

```
GET /path/to/resource?date=YYYYMMDD&other=params
```

### Response

The proxy forwards the complete request (headers, body, query params) to the appropriate downstream service based on the date parameter.

### Error Responses

- `400 Bad Request`: Missing or invalid date parameter
- `404 Not Found`: No service configured for the given date
- `500 Internal Server Error`: Proxy or downstream service error

## Security Features

- **Non-root execution**: Containers run as non-root user
- **Read-only filesystem**: Root filesystem is read-only
- **Security contexts**: Proper security contexts with capability dropping
- **Resource limits**: CPU and memory limits defined
- **Health checks**: Liveness and readiness probes

## Monitoring

The proxy includes:
- Structured logging with request details
- HTTP client with connection pooling and timeouts
- Graceful shutdown with configurable timeout
- X-Forwarded-* headers for request tracing

## Development

### Project Structure

```
dateproxy/
├── main.go              # Application entry point
├── proxy.go             # Core proxy logic
├── config.go            # Configuration management
├── config.yaml          # Default configuration
├── Dockerfile           # Multi-stage Docker build
├── Makefile             # Build and development tasks
├── k8s/                 # Kubernetes manifests
│   ├── configmap.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── backend-services.yaml
└── test-backends/       # Test services
    ├── service1/
    ├── service2/
    └── service3/
```

### Building

```bash
# Build proxy
make build

# Build all services
make build-services

# Clean artifacts
make clean
```

### Testing

The included test services return JSON responses showing:
- Which service handled the request
- Request details (path, method, headers, query params)
- Timestamp

This allows easy verification of routing behavior.