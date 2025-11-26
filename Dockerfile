# Stage 1: Build
FROM golang:1.25.3-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build static binary
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -tags kubernetes -ldflags="-s -w" -o /app/discovery ./cmd/discovery

# Stage 2: Minimal runtime
FROM scratch
# CA certificates for TLS (optional)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/discovery /discovery
COPY --from=builder /src/configuration.yaml /configuration.yaml

ENTRYPOINT ["/discovery"]
