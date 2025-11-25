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
	Listen                []string              `json:"listen"`
	Routes                []Route               `json:"routes"`
	TLSConnectionPolicies []TLSConnectionPolicy `json:"tls_connection_policies,omitempty"`
}

type TLSConnectionPolicy struct {
	Certificate *Certificate `json:"certificate,omitempty"`
}

type Certificate struct {
	CertificateFile string `json:"certificate_file,omitempty"`
	KeyFile         string `json:"key_file,omitempty"`
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

	// optional transport configuration for reverse_proxy upstreams
	Transport *Transport `json:"transport,omitempty"`
}

type Upstream struct {
	Dial string `json:"dial"`
}

type Transport struct {
	Protocol string        `json:"protocol,omitempty"`
	TLS      *TransportTLS `json:"tls,omitempty"`
}

type TransportTLS struct {
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`
}

type EndpointInfo struct {
	Port     int    `yaml:"port"`
	Domain   string `yaml:"domain"`
	Upstream string `yaml:"upstream"`
}

type LifecycleEvent struct {
	ContainerInfo EndpointInfo
	EventType     EventType
}

type ProviderConnector interface {
	GetRoutes() ([]Route, error)
	GetEventChannel() <-chan LifecycleEvent
}

type TLSConfig struct {
	Manual       bool   `mapstructure:"manual"`
	CertFilePath string `mapstructure:"certFilePath"`
	KeyFilePath  string `mapstructure:"keyFilePath"`
}

type ManualRoute struct {
	Domain   string `yaml:"domain"`
	Upstream string `yaml:"upstreamUrl"`
	TLS      bool   `yaml:"tls"`
}

type CaddyConfig struct {
	ManualRoutes  []ManualRoute `yaml:"routes"`
	TLSConfig     TLSConfig
	CaddyAdminUrl string
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
