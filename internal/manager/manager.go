package manager

import (
	"fmt"
	"log"

	"github.com/jaku01/caddyservicediscovery/internal/caddy"
)

func StartServiceDiscovery(caddyAdminUrl string, providerConnector caddy.ProviderConnector) error {
	caddyConnector := caddy.NewConnector(caddyAdminUrl)

	log.Println("Starting manager for service discovery")
	log.Printf("Using caddy admin url: %s", caddyAdminUrl)

	err := createCaddyConfigIfMissing(caddyConnector)
	if err != nil {
		return err
	}

	routes, err := configureInitialRoutes(err, providerConnector, caddyConnector)
	if err != nil {
		return err
	}

	err = handleLifecycleEvents(providerConnector, err, routes, caddyConnector)
	if err != nil {
		return err
	}
	return nil
}

func handleLifecycleEvents(providerConnector caddy.ProviderConnector, err error, routes []caddy.Route, caddyConnector *caddy.Connector) error {
	for lifecycleEvent := range providerConnector.GetEventChannel() {
		log.Printf("Received lifecycle event %+v\n", lifecycleEvent)
		err = updateRoutes(lifecycleEvent, &routes)
		if err != nil {
			return err
		}

		routes = ensureFallbackAtEnd(routes)
		err = caddyConnector.SetRoutes(routes)
		if err != nil {
			return err
		}
	}
	return nil
}

func configureInitialRoutes(err error, providerConnector caddy.ProviderConnector, caddyConnector *caddy.Connector) ([]caddy.Route, error) {
	routes, err := providerConnector.GetRoutes()
	if err != nil {
		return nil, err
	}
	log.Println("Initial server map retrieved, updating caddy configuration")

	_ = caddyConnector.PrintCurrentConfig()

	if !fallbackExists(routes) {
		routes = append(routes, caddy.New404FallbackRoute())
	}
	err = caddyConnector.SetRoutes(routes)
	if err != nil {
		return nil, err
	}
	return routes, nil
}

func updateRoutes(lifecycleEvent caddy.LifecycleEvent, routes *[]caddy.Route) error {
	switch lifecycleEvent.EventType {
	case caddy.StartEvent:
		fmt.Printf("Adding route %+v\n", lifecycleEvent.ContainerInfo)
		// Deduplicate
		for _, r := range *routes {
			if sameRoute(r, lifecycleEvent.ContainerInfo) {
				return nil
			}
		}
		*routes = append(*routes, caddy.NewReverseProxyRoute(
			lifecycleEvent.ContainerInfo.Domain,
			lifecycleEvent.ContainerInfo.Upstream,
		))
		return nil

	case caddy.DieEvent:
		fmt.Printf("Removing route %+v\n", lifecycleEvent.ContainerInfo)
		newRoutes := make([]caddy.Route, 0, len(*routes))
		removed := false
		for _, r := range *routes {
			if sameRoute(r, lifecycleEvent.ContainerInfo) {
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

func sameRoute(r caddy.Route, info caddy.ContainerInfo) bool {
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

func fallbackExists(routes []caddy.Route) bool {
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

func ensureFallbackAtEnd(routes []caddy.Route) []caddy.Route {
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

func createCaddyConfigIfMissing(caddyConnector *caddy.Connector) error {
	config, err := caddyConnector.GetCaddyConfig()
	if err != nil && err.Error() != "no caddy config found" {
		return err
	}
	if config != nil {
		return nil
	}

	fmt.Println("No caddy config found, creating one")
	err = caddyConnector.CreateCaddyConfig()
	if err != nil {
		return err
	}
	return err
}
