package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Connector struct {
	Url       string
	TlsConfig TLSConfig
}

func NewConnector(url string, tlsConfig TLSConfig) *Connector {
	return &Connector{
		Url:       url,
		TlsConfig: tlsConfig,
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

	// if the content is "null", return nil
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
	config.Apps.HTTP.Servers = make(map[string]Server, 1)

	server := Server{
		Listen: []string{":443", ":80"},
		Routes: []Route{},
	}

	if c.TlsConfig.Manual {
		slog.Info("Using manual TLS configuration",
			"certFilePath", c.TlsConfig.CertFilePath,
			"keyFilePath", c.TlsConfig.KeyFilePath)
		server.TLSConnectionPolicies = []TLSConnectionPolicy{
			{
				Certificate: &Certificate{
					CertificateFile: c.TlsConfig.CertFilePath,
					KeyFile:         c.TlsConfig.KeyFilePath,
				},
			},
		}
	}

	config.Apps.HTTP.Servers["srv0"] = server

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

func (c *Connector) SetRoutes(routes []Route) error {
	reqBody, err := json.Marshal(routes)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, c.Url+"/config/apps/http/servers/srv0/routes/", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// NewReverseProxyRoute creates a reverse proxy forwarding accesses to incomingDomain to upstreamPort
func NewReverseProxyRoute(incomingDomain string, upstream string) Route {
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
										Dial: upstream,
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
	converted, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Config: %s\n", string(converted))
	return nil
}
