//go:build docker

package main

import (
	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	"github.com/jaku01/caddyservicediscovery/internal/docker"
)

// Returns a Docker connector when built with the 'docker' tag.
func newServiceDiscoveryProviderConnector() (caddy.ServiceDiscoveryProvider, error) {
	return docker.NewDockerConnector(), nil
}
