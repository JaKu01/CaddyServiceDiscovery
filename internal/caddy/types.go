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
}

type Upstream struct {
	Dial string `json:"dial"`
}
