# Caddy Service Discovery

This project provides automated service discovery for [Caddy](https://caddyserver.com/) using Docker container labels. It dynamically updates the Caddy HTTP server configuration based on running Docker containers, enabling seamless reverse proxying for services as they start and stop.

## Features

- **Automatic Discovery:** Detects Docker containers with specific labels and manages their reverse proxy configuration in Caddy.
- **Dynamic Updates:** Monitors containers and updates Caddy only when changes are detected.
- **Zero Downtime:** Updates Caddy configuration without restarting the server.
- **Easy Integration:** Works with any Dockerized service by adding the required labels.

## How It Works

1. The scheduler periodically queries Docker for containers with the label `caddy.service.discovery.active=true`.
2. For each matching container, it reads the labels:
    - `caddy.service.discovery.port`: The port to expose via Caddy.
    - `caddy.service.discovery.upstream`: The upstream address for reverse proxying.
3. It generates a Caddy server configuration for each container.
4. If the set of active containers changes, it updates the Caddy configuration via the Caddy Admin API.

## Getting Started

### Prerequisites

- [Go](https://golang.org/) (for building the tool)
- [Docker](https://www.docker.com/)
- [Caddy](https://caddyserver.com/) (running with the Admin API enabled, default port 2019)

### Build

```sh
go build -o caddyservicediscovery ./cmd/discovery
```

### Usage

Start the Caddy server (ensure the Admin API is accessible):

```sh
caddy run
```

Start the service discovery tool:

```sh
./caddyservicediscovery
```

By default, it connects to `http://localhost:2019` for the Caddy Admin API.

### Label Your Containers

When running your Docker containers, add the following labels:

- `caddy.service.discovery.active=true`
- `caddy.service.discovery.port=<port>` (the port to expose)
- `caddy.service.discovery.upstream=<host:port>` (the upstream address)

**Example:**

```sh
docker run -d \
  -p 7123:7123 \
  --name my-service \
  --label caddy.service.discovery.upstream=localhost:7123 \
  --label caddy.service.discovery.port=3080 \
  --label caddy.service.discovery.active=true \
  my-image:latest
```

## Configuration File (`configuration.yaml`)

You can configure the service discovery tool using a `configuration.yaml` file in the project root. The following options are available:

- `CaddyAdminUrl`: The URL of the Caddy Admin API. Default is `http://localhost:2019`.

**Example:**

```yaml
CaddyAdminUrl: "http://localhost:2019"
```

This allows you to easily adjust the connection to your Caddy instance and how frequently the service discovery runs, without changing the code.

## Project Structure

- `cmd/discovery/main.go`: Entry point for the service discovery tool.
- `internal/caddy/`: Handles Caddy API communication and configuration.
- `internal/docker/`: Handles Docker API communication and container discovery.
- `internal/scheduler/`: Orchestrates the discovery and update loop.

## Development

- Clone the repository.
- Run `go mod tidy` to install dependencies.
- Build and run as described above.


## License

MIT License

---

For questions or contributions, please open an issue or pull request.