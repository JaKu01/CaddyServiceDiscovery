package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/jaku01/caddyservicediscovery/internal/discovery"
)

type Connector struct {
	Config *discovery.CaddyConfig
}

func NewConnector(caddyConfig discovery.CaddyConfig) *Connector {
	return &Connector{
		Config: &caddyConfig,
	}
}

func (c *Connector) GetCaddyConfig() (*Config, error) {
	url := c.Config.CaddyAdminUrl + "/config/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request to %s failed with status code %d", url, resp.StatusCode)
	}

	responseContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// if the content is "null", return nil
	if len(responseContent) == 0 || string(responseContent) == "null\n" {
		return nil, fmt.Errorf("no caddy Config found")
	}

	caddyConfig, err := UnmarshalCaddyConfig(responseContent)
	if err != nil {
		return nil, err
	}
	return &caddyConfig, nil
}

func (c *Connector) CreateCaddyConfig() error {
	config := Config{}
	config.Apps.HTTP.Servers = make(map[string]Server, 1)

	server := Server{
		Listen: []string{":443", ":80"},
		Routes: []Route{},
	}

	config.Apps.HTTP.Servers["srv0"] = server

	if c.Config.TLSConfig.Manual {
		slog.Info("Using manual TLS configuration",
			"certFilePath", c.Config.TLSConfig.CertFilePath,
			"keyFilePath", c.Config.TLSConfig.KeyFilePath)

		config.Apps.TLS = &TLSApp{
			Certificates: Certificates{
				LoadFiles: []LoadFile{
					{
						Certificate: c.Config.TLSConfig.CertFilePath,
						Key:         c.Config.TLSConfig.KeyFilePath,
					},
				},
			},
		}
	}

	url := c.Config.CaddyAdminUrl + "/load"
	bodyContent, err := json.Marshal(config)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyContent))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request to %s failed with status code %d", url, resp.StatusCode)
	}

	slog.Info("Created Caddy config successfully")
	return nil
}

func (c *Connector) SetRoutes(routes []Route) error {
	reqBody, err := json.Marshal(routes)
	if err != nil {
		return err
	}

	url := c.Config.CaddyAdminUrl + "/config/apps/http/servers/srv0/routes/"
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	return nil
}

// NewReverseProxyRoute creates a reverse proxy forwarding accesses to incomingDomain to upstreamPort
func NewReverseProxyRoute(incomingDomain string, upstreamAddr string) Route {
	return Route{
		Handle: []Handle{
			{
				Handler: "subroute",
				Routes: []Route{
					{
						Match: nil,
						Handle: []Handle{
							{
								Handler: "reverse_proxy",
								Upstreams: []Upstream{
									{
										Dial: upstreamAddr,
									},
								},
							},
						},
					},
				},
			},
		},
		Match: []Match{
			{
				Host: []string{incomingDomain},
			},
		},
	}
}

func NewExternalReverseProxyRoute(incomingDomain string, upstream string, tls bool) Route {
	upstreamHandle := Handle{
		Handler: "reverse_proxy",
		Upstreams: []Upstream{
			{
				Dial: upstream,
			},
		},
	}

	if tls {
		upstreamHandle.Transport = &Transport{
			Protocol: "http",
			TLS:      &TransportTLS{},
		}
	}

	return Route{
		Handle: []Handle{
			{
				Handler: "subroute",
				Routes: []Route{
					{
						Match: nil,
						Handle: []Handle{
							upstreamHandle,
						},
					},
				},
			},
		},
		Match: []Match{
			{
				Host: []string{incomingDomain},
			},
		},
	}
}

func New404FallbackRoute() Route {
	return Route{
		Match: []Match{{}}, // match everything
		Handle: []Handle{
			{
				Handler:    "static_response",
				StatusCode: 404,
				Body:       "Not Found",
			},
		},
	}
}

func (c *Connector) PrintCurrentConfig() error {
	config, err := c.GetCaddyConfig()
	if err != nil {
		return err
	}

	converted, err := json.Marshal(config)
	if err != nil {
		return err
	}
	fmt.Printf("Config: %s\n", string(converted))
	return nil
}
