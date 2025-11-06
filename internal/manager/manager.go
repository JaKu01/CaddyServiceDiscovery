package manager

import (
	"fmt"
	"log"

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

	serverMap, err := getServerMap(dockerConnector)
	if err != nil {
		return err
	}
	log.Println("Initial server map retrieved, updating caddy configuration")

	err = caddyConnector.SetServers(serverMap)
	if err != nil {
		return err
	}

	for dockerEvent := range dockerConnector.GetEventChannel() {
		log.Printf("Received docker event %+v, updating caddy configuration\n", dockerEvent)
		err = updateServerMap(dockerEvent, serverMap)
		if err != nil {
			return err
		}

		err = caddyConnector.SetServers(serverMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func getServerMap(dockerConnector *dockerconnector.DockerConnector) (map[string]caddy.Server, error) {
	containers, err := dockerConnector.GetAllContainersWithActiveLabel()
	if err != nil {
		return nil, err
	}

	serverMap := make(map[string]caddy.Server)
	for _, container := range containers {
		reverseProxyServer := caddy.NewReverseProxyServer(container.Port, container.Upstream)
		serverMap[container.ContainerName] = reverseProxyServer
	}
	return serverMap, nil
}

func updateServerMap(dockerEvent dockerconnector.DockerEvent, serverMap map[string]caddy.Server) error {
	switch dockerEvent.EventType {
	case dockerconnector.ContainerStartEvent:
		reverseProxyServer := caddy.NewReverseProxyServer(dockerEvent.ContainerInfo.Port, dockerEvent.ContainerInfo.Upstream)
		serverMap[dockerEvent.ContainerInfo.ContainerName] = reverseProxyServer
	case dockerconnector.ContainerDieEvent:
		delete(serverMap, dockerEvent.ContainerInfo.ContainerName)
	default:
		return fmt.Errorf("unknown docker event type %d", dockerEvent.EventType)
	}
	return nil
}

func createCaddyConfigIfMissing(caddyConnector *caddy.Connector) error {
	config, err := caddyConnector.GetCaddyConfig()
	if err != nil && err.Error() != "no caddy config found" {
		panic(err)
	}
	if config != nil {
		return nil
	}

	fmt.Println("No caddy config found, creating one")
	err = caddyConnector.CreateCaddyConfig()
	if err != nil {
		panic(err)
	}
	return err
}
