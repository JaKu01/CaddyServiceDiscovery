package provider

import "github.com/jaku01/caddyservicediscovery/internal/caddy"

type ServiceDiscoveryProvider interface {
	GetRoutes() ([]caddy.Route, error)
	GetEventChannel() <-chan LifecycleEvent
}

type EndpointInfo struct {
	Port     int    `yaml:"port"`
	Domain   string `yaml:"domain"`
	Upstream string `yaml:"upstream"`
}

type LifecycleEvent struct {
	ContainerInfo      EndpointInfo
	LifeCycleEventType EventType
}

type EventType int

const (
	StartEvent = iota
	DieEvent
)

func (e EventType) String() string {
	switch e {
	case StartEvent:
		return "StartEvent"
	case DieEvent:
		return "DieEvent"
	default:
		return "unknown"
	}
}
