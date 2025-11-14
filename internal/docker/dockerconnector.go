package dockerconnector

import (
	"context"
	"log"
	"strconv"

	containertypes "github.com/docker/docker/api/types/container"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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

type ContainerInfo struct {
	Port          int
	Domain        string
	ContainerName string
}

type EventType int

const (
	ContainerStartEvent = iota
	ContainerDieEvent
)

func (e EventType) String() string {
	switch e {
	case ContainerStartEvent:
		return "ContainerStartEvent"
	case ContainerDieEvent:
		return "ContainerDieEvent"
	default:
		return "unknown"
	}
}

type DockerEvent struct {
	ContainerInfo ContainerInfo
	EventType     EventType
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

func (dc *DockerConnector) GetEventChannel() <-chan DockerEvent {
	transformedEvents := make(chan DockerEvent)

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
				log.Println("Error listening to docker events:", err)
			case <-dc.ctx.Done():
				return
			}
		}
	}()

	return transformedEvents
}

func transformDockerEvent(rawEvent eventtypes.Message) *DockerEvent {
	if rawEvent.Type != eventtypes.ContainerEventType || rawEvent.Actor.Attributes[activeLabel] != "true" {
		return nil
	}

	var eventType EventType
	switch rawEvent.Action {
	case eventtypes.ActionStart:
		eventType = ContainerStartEvent
	case eventtypes.ActionDie:
		eventType = ContainerDieEvent
	default:
		return nil
	}

	portStr := rawEvent.Actor.Attributes[portLabel]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Error converting docker container port '%s' to int\n", portStr)
		return nil
	}

	containerInfo := ContainerInfo{
		Port:          port,
		Domain:        rawEvent.Actor.Attributes[domainLabel],
		ContainerName: rawEvent.Actor.Attributes["name"],
	}

	return &DockerEvent{
		ContainerInfo: containerInfo,
		EventType:     eventType,
	}
}

func (dc *DockerConnector) GetAllContainersWithActiveLabel() ([]ContainerInfo, error) {
	containers, err := dc.dockerClient.ContainerList(dc.ctx, containertypes.ListOptions{})
	if err != nil {
		return nil, err
	}

	var activeContainers []ContainerInfo
	for _, container := range containers {
		if container.Labels[activeLabel] == "true" {
			port, err := strconv.Atoi(container.Labels[portLabel])
			if err != nil {
				log.Println("Error converting port to int")
				continue
			}

			containerInfo := ContainerInfo{
				Port:          port,
				Domain:        container.Labels[domainLabel],
				ContainerName: container.Names[0],
			}

			activeContainers = append(activeContainers, containerInfo)
		}
	}

	return activeContainers, nil
}
