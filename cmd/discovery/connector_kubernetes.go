//go:build kubernetes

package main

import (
	"github.com/jaku01/caddyservicediscovery/internal/provider"
	"github.com/jaku01/caddyservicediscovery/internal/provider/kubernetes"
)

// Returns a Kubernetes connector when built with the 'kubernetes' tag.
func newServiceDiscoveryProviderConnector() (provider.ServiceDiscoveryProvider, error) {
	return kubernetes.NewKubernetesConnector()
}
