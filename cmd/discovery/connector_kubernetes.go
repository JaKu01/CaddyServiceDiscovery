//go:build kubernetes

package main

import (
	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	"github.com/jaku01/caddyservicediscovery/internal/kubernetes"
)

// Returns a Kubernetes connector when built with the 'kubernetes' tag.
func newProviderConnector() (caddy.ServiceDiscoveryProvider, error) {
	return kubernetes.NewKubernetesConnector()
}
