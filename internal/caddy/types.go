package caddy

import (
	"encoding/json"
)

func UnmarshalCaddyConfig(data []byte) (Config, error) {
	var r Config
	err := json.Unmarshal(data, &r)
	return r, err
}

type Config struct {
	Apps struct {
		HTTP struct {
			Servers map[string]Server `json:"servers"`
		} `json:"http"`
	} `json:"apps"`
}

type Server struct {
	Listen []string `json:"listen"`
	Routes []Route  `json:"routes"`
}

type Route struct {
	Match  []Match  `json:"match,omitempty"`
	Handle []Handle `json:"handle"`
}

type Match struct {
	Host []string `json:"host,omitempty"`
}

type Handle struct {
	Handler   string     `json:"handler"`
	Routes    []Route    `json:"routes,omitempty"`
	Upstreams []Upstream `json:"upstreams,omitempty"`

	// for static_response
	StatusCode int    `json:"status_code,omitempty"`
	Body       string `json:"body,omitempty"`
}

type Upstream struct {
	Dial string `json:"dial"`
}

type ContainerInfo struct {
	Port     int
	Domain   string
	Upstream string
}

type LifecycleEvent struct {
	ContainerInfo ContainerInfo
	EventType     EventType
}

type ProviderConnector interface {
	GetRoutes() ([]Route, error)
	GetEventChannel() <-chan LifecycleEvent
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
