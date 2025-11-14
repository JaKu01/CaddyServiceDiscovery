package manager

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	dockerconnector "github.com/jaku01/caddyservicediscovery/internal/docker"
)

func StartServiceDiscovery(caddyAdminUrl string) error {
	dockerConnector := dockerconnector.NewDockerConnector()
	caddyConnector := caddy.NewConnector(caddyAdminUrl)

	log.Println("Starting manager for service discovery")
	log.Printf("Using caddy admin url: %s", caddyAdminUrl)

	err := createCaddyConfigIfMissing(caddyConnector)
	if err != nil {
		return err
	}

	routes, err := getRoutes(dockerConnector)
	if err != nil {
		return err
	}
	log.Println("Initial server map retrieved, updating caddy configuration")

	err = caddyConnector.SetRoutes(routes)
	if err != nil {
		return err
	}

	for dockerEvent := range dockerConnector.GetEventChannel() {
		log.Printf("Received docker event %+v, updating caddy configuration\n", dockerEvent)
		err = updateRoutes(dockerEvent, &routes)
		if err != nil {
			return err
		}

		err = caddyConnector.SetRoutes(routes)
		if err != nil {
			return err
		}
	}

	return nil
}

func getRoutes(dockerConnector *dockerconnector.DockerConnector) ([]caddy.Route, error) {
	containers, err := dockerConnector.GetAllContainersWithActiveLabel()
	if err != nil {
		return nil, err
	}

	routes := make([]caddy.Route, 0, len(containers))
	for _, container := range containers {
		reverseProxyRoute := caddy.NewReverseProxyRoute(container.Domain, container.Port)
		routes = append(routes, reverseProxyRoute)
	}
	return routes, nil
}

func updateRoutes(dockerEvent dockerconnector.DockerEvent, routes *[]caddy.Route) error {
	switch dockerEvent.EventType {
	case dockerconnector.ContainerStartEvent:
		reverseProxyRoute := caddy.NewReverseProxyRoute(dockerEvent.ContainerInfo.Domain, dockerEvent.ContainerInfo.Port)
		*routes = append(*routes, reverseProxyRoute)
	case dockerconnector.ContainerDieEvent:
		oldLen := len(*routes)
		for i, route := range *routes {
			portMatches := route.Handle[0].Routes[0].Handle[0].Upstreams[0].Dial == ":"+strconv.Itoa(dockerEvent.ContainerInfo.Port)
			domainMatches := route.Match[0].Host[0] == dockerEvent.ContainerInfo.Domain
			if portMatches && domainMatches {
				// Delete entry, see https://go.dev/wiki/SliceTricks#delete
				*routes = append((*routes)[:i], (*routes)[i+1:]...)
			}
		}
		newLen := len(*routes)
		if newLen-1 != oldLen {
			return fmt.Errorf("route to be removed for docker event %#v not found\n", dockerEvent)
		}
	default:
		return fmt.Errorf("unknown docker event type %d", dockerEvent.EventType)
	}
	return nil
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
