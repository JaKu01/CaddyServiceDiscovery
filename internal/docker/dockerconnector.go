package docker

import (
	"context"
	"log/slog"
	"strconv"

	containertypes "github.com/docker/docker/api/types/container"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/jaku01/caddyservicediscovery/internal/caddy"
)

const (
	activeLabel = "caddy.service.discovery.active"
	portLabel   = "caddy.service.discovery.port"
	domainLabel = "caddy.service.discovery.domain"
)

type DockerConnector struct {
	dockerClient *client.Client
	ctx          context.Context
}

func NewDockerConnector() *DockerConnector {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &DockerConnector{
		dockerClient: cli,
		ctx:          ctx,
	}
}

func (dc *DockerConnector) GetRoutes() ([]caddy.Route, error) {
	containers, err := dc.GetAllContainersWithActiveLabel()
	if err != nil {
		return nil, err
	}

	routes := make([]caddy.Route, 0, len(containers))
	for _, container := range containers {
		reverseProxyRoute := caddy.NewReverseProxyRoute(container.Domain, container.Upstream)
		routes = append(routes, reverseProxyRoute)
	}
	return routes, nil
}

func (dc *DockerConnector) GetEventChannel() <-chan caddy.LifecycleEvent {
	transformedEvents := make(chan caddy.LifecycleEvent)

	go func() {
		defer close(transformedEvents)
		ctxWithCancel, cancelCtx := context.WithCancel(dc.ctx)
		// For available filters, see https://docs.docker.com/reference/api/engine/version/v1.51/#tag/System/operation/SystemEvents
		eventFilters := filters.NewArgs()
		eventFilters.Add("type", string(eventtypes.ContainerEventType))
		eventFilters.Add("event", string(eventtypes.ActionStart))
		eventFilters.Add("event", string(eventtypes.ActionDie))
		eventFilters.Add("label", activeLabel)
		rawEvents, err := dc.dockerClient.Events(ctxWithCancel, eventtypes.ListOptions{Filters: eventFilters})
		defer cancelCtx()

		for {
			select {
			case event, ok := <-rawEvents:
				if !ok {
					return
				}
				transformedEvent := transformDockerEvent(event)
				if transformedEvent == nil {
					continue
				}
				transformedEvents <- *transformedEvent
			case err := <-err:
				slog.Error("Error listening to docker events", "error", err)
			case <-dc.ctx.Done():
				return
			}
		}
	}()

	return transformedEvents
}

func transformDockerEvent(rawEvent eventtypes.Message) *caddy.LifecycleEvent {
	if rawEvent.Type != eventtypes.ContainerEventType || rawEvent.Actor.Attributes[activeLabel] != "true" {
		return nil
	}

	var eventType caddy.EventType
	switch rawEvent.Action {
	case eventtypes.ActionStart:
		eventType = caddy.StartEvent
	case eventtypes.ActionDie:
		eventType = caddy.DieEvent
	default:
		return nil
	}

	portStr := rawEvent.Actor.Attributes[portLabel]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		slog.Error("Error converting docker container port to int", "port", portStr)
		return nil
	}

	containerInfo := caddy.ContainerInfo{
		Port:     port,
		Domain:   rawEvent.Actor.Attributes[domainLabel],
		Upstream: ":" + portStr,
	}

	return &caddy.LifecycleEvent{
		ContainerInfo: containerInfo,
		EventType:     eventType,
	}
}

func (dc *DockerConnector) GetAllContainersWithActiveLabel() ([]caddy.ContainerInfo, error) {
	containers, err := dc.dockerClient.ContainerList(dc.ctx, containertypes.ListOptions{})
	if err != nil {
		return nil, err
	}

	var activeContainers []caddy.ContainerInfo
	for _, container := range containers {
		if container.Labels[activeLabel] == "true" {
			port, err := strconv.Atoi(container.Labels[portLabel])
			if err != nil {
				slog.Error("Error converting port to int")
				continue
			}

			containerInfo := caddy.ContainerInfo{
				Port:     port,
				Domain:   container.Labels[domainLabel],
				Upstream: ":" + container.Labels[portLabel],
			}

			activeContainers = append(activeContainers, containerInfo)
		}
	}

	return activeContainers, nil
}
