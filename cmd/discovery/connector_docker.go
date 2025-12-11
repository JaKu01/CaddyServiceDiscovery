//go:build docker

package main

import (
	"github.com/jaku01/caddyservicediscovery/internal/provider"
	"github.com/jaku01/caddyservicediscovery/internal/provider/docker"
)

// Returns a Docker connector when built with the 'docker' tag.
func newServiceDiscoveryProviderConnector() (provider.ServiceDiscoveryProvider, error) {
	return docker.NewDockerConnector(), nil
}
