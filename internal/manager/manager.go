package manager

import (
	"fmt"
	"log/slog"

	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	"github.com/jaku01/caddyservicediscovery/internal/discovery"
	"github.com/jaku01/caddyservicediscovery/internal/provider"
)

func StartServiceDiscovery(caddyConnector *caddy.Connector, providerConnector provider.ServiceDiscoveryProvider) error {
	slog.Info("Starting manager for service discovery")
	slog.Info("Using caddy admin api", "url", caddyConnector.Config.CaddyAdminUrl)

	err := caddyConnector.CreateCaddyConfig()
	if err != nil {
		return err
	}

	routes, err := configureInitialRoutes(providerConnector, caddyConnector)
	if err != nil {
		return err
	}

	err = handleLifecycleEvents(providerConnector, routes, caddyConnector)
	if err != nil {
		return err
	}
	return nil
}

func handleLifecycleEvents(providerConnector provider.ServiceDiscoveryProvider, routes []caddy.Route, caddyConnector *caddy.Connector) error {
	for lifecycleEvent := range providerConnector.GetEventChannel() {
		slog.Info("Received lifecycle event", "content", lifecycleEvent)
		err := updateRoutes(lifecycleEvent, &routes)
		if err != nil {
			return err
		}

		routes = ensureFallbackRoute(routes)
		err = caddyConnector.SetRoutes(routes)
		if err != nil {
			return err
		}
	}
	return nil
}

func configureInitialRoutes(providerConnector provider.ServiceDiscoveryProvider, caddyConnector *caddy.Connector) ([]caddy.Route, error) {
	routes, err := providerConnector.GetRoutes()
	if err != nil {
		return nil, err
	}
	slog.Info("Initial server map retrieved, updating caddy configuration")

	for _, manualRoute := range caddyConnector.Config.ManualRoutes {
		if !hasManualRoute(routes, manualRoute) {
			reverseProxyRoute := caddy.NewExternalReverseProxyRoute(manualRoute.Domain, manualRoute.Upstream, manualRoute.TLS)
			routes = append(routes, reverseProxyRoute)
			slog.Info("Added manual route", "route", manualRoute)
		}
	}

	if !fallbackRouteExists(routes) {
		routes = append(routes, caddy.New404FallbackRoute())
	}
	err = caddyConnector.SetRoutes(routes)
	if err != nil {
		return nil, err
	}
	_ = caddyConnector.PrintCurrentConfig()
	return routes, nil
}

func updateRoutes(lifecycleEvent provider.LifecycleEvent, routes *[]caddy.Route) error {
	switch lifecycleEvent.LifeCycleEventType {
	case provider.StartEvent:
		slog.Info("Adding route", "detail", lifecycleEvent.ContainerInfo)
		// Deduplicate
		for _, r := range *routes {
			if isSameRoute(r, lifecycleEvent.ContainerInfo) {
				return nil
			}
		}
		*routes = append(*routes, caddy.NewReverseProxyRoute(
			lifecycleEvent.ContainerInfo.Domain,
			lifecycleEvent.ContainerInfo.Upstream,
		))
		return nil

	case provider.DieEvent:
		slog.Info("Removing route", "detail", lifecycleEvent.ContainerInfo)
		newRoutes := make([]caddy.Route, 0, len(*routes))
		removed := false
		for _, r := range *routes {
			if isSameRoute(r, lifecycleEvent.ContainerInfo) {
				removed = true
				continue
			}
			newRoutes = append(newRoutes, r)
		}
		if !removed {
			return fmt.Errorf("route not found for %+v", lifecycleEvent)
		}
		*routes = newRoutes
		return nil
	}

	return fmt.Errorf("unknown lifecycle event")
}

func isSameRoute(r caddy.Route, info provider.EndpointInfo) bool {
	// Host
	if len(r.Match) == 0 || len(r.Match[0].Host) == 0 {
		return false
	}
	if r.Match[0].Host[0] != info.Domain {
		return false
	}

	// Upstream
	hs := r.Handle[0]
	rp := hs.Routes[0].Handle[0]
	if len(rp.Upstreams) == 0 {
		return false
	}

	return rp.Upstreams[0].Dial == info.Upstream
}

func hasManualRoute(routes []caddy.Route, route discovery.ManualRoute) bool {
	for _, r := range routes {
		if r.Match[0].Host[0] == route.Domain {
			return true
		}
	}
	return false
}

func fallbackRouteExists(routes []caddy.Route) bool {
	for _, r := range routes {
		if len(r.Handle) == 0 {
			continue
		}
		h := r.Handle[0]
		if h.Handler == "static_response" && h.StatusCode == 404 {
			return true
		}
	}
	return false
}

func ensureFallbackRoute(routes []caddy.Route) []caddy.Route {
	var fallback *caddy.Route
	filtered := make([]caddy.Route, 0, len(routes))

	// Remove the fallback route
	for _, r := range routes {
		if len(r.Handle) > 0 && r.Handle[0].Handler == "static_response" && r.Handle[0].StatusCode == 404 {
			fallback = &r
		} else {
			filtered = append(filtered, r)
		}
	}

	// Add the fallback route to the end of the handlers
	if fallback != nil {
		filtered = append(filtered, *fallback)
	}

	return filtered
}
