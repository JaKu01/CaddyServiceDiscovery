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
	config.Apps.HTTP.Servers = make(map[string]Server, 1)
	config.Apps.HTTP.Servers["srv0"] = Server{
		Listen: []string{":443", ":80"},
		Routes: []Route{},
	}
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

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.Status, string(respBody))

	return nil
}

// NewReverseProxyRoute creates a reverse proxy forwarding accesses to incomingDomain to upstreamPort
func NewReverseProxyRoute(incomingDomain string, upstreamPort int) Route {
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
										Dial: ":" + strconv.Itoa(upstreamPort),
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
