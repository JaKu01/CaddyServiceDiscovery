package caddy

import "encoding/json"

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

		TLS *TLSApp `json:"tls,omitempty"`
	} `json:"apps"`
}

type Server struct {
	Listen                []string              `json:"listen"`
	Routes                []Route               `json:"routes"`
	TLSConnectionPolicies []TLSConnectionPolicy `json:"tls_connection_policies,omitempty"`
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

type TLSConnectionPolicy struct {
}

type TLSApp struct {
	Certificates Certificates `json:"certificates"`
}

type Certificates struct {
	LoadFiles []LoadFile `json:"load_files,omitempty"`
}

type LoadFile struct {
	Certificate string `json:"certificate"`
	Key         string `json:"key"`
}
