package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Connector struct {
	Url string
}

func NewConnector(url string) *Connector {
	return &Connector{
		Url: url,
	}
}

func (c *Connector) GetCaddyConfig() (*Config, error) {
	resp, err := http.Get(c.Url + "/config/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// if content is "null", return nil
	if len(responseContent) == 0 || string(responseContent) == "null\n" {
		return nil, fmt.Errorf("no caddy config found")
	}

	caddyConfig, err := UnmarshalCaddyConfig(responseContent)
	if err != nil {
		return nil, err
	}
	return &caddyConfig, nil
}

func (c *Connector) CreateCaddyConfig() error {
	config := Config{}
	body, err := json.Marshal(config)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Url+"/load", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Connector) SetServers(servers map[string]Server) error {
	body, err := json.Marshal(servers)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Url+"/config/apps/http/servers", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func NewReverseProxyServer(port int, upstream string) Server {
	server := Server{
		Listen: []string{":" + strconv.Itoa(port)},
		Routes: []Route{
			{
				Handle: []Handler{
					{
						Handler: "reverse_proxy",
						Upstreams: []Upstream{
							{Dial: upstream},
						},
					},
				},
			},
		},
	}
	return server
}
