package caddy

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestConnector_GetCaddyConfig(t *testing.T) {
	mockResponse := "{\"apps\":{\"http\":{\"servers\":{\"exampleServer\":{\"listen\":[\":443\"],\"routes\":[{\"handle\":[{\"handler\":\"reverse_proxy\",\"upstreams\":[{\"dial\":\":8080\"}]}]}]}}}}}\n"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(mockResponse))
		} else {
			t.Errorf("Expected %s with method %s, got %s with method %s",
				"/config/", http.MethodGet, r.URL.Path, r.Method)
		}
	}))

	connector := NewConnector(mockServer.URL)

	config, err := connector.GetCaddyConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	server, ok := config.Apps.HTTP.Servers["exampleServer"]
	if !ok {
		t.Fatalf("Expected server for exampleServer")
	}
	if server.Listen[0] != ":443" {
		t.Errorf("Expected listen port :443, got %s", server.Listen[0])
	}
	if len(server.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(server.Routes))
	}
	if len(server.Routes[0].Handle) != 1 {
		t.Errorf("Expected 1 handle, got %d", len(server.Routes[0].Handle))
	}
	if server.Routes[0].Handle[0].Handler != "reverse_proxy" {
		t.Errorf("Expected handler reverse_proxy, got %s", server.Routes[0].Handle[0].Handler)
	}
	if len(server.Routes[0].Handle[0].Upstreams) != 1 {
		t.Errorf("Expected 1 upstream, got %d", len(server.Routes[0].Handle[0].Upstreams))
	}
	if server.Routes[0].Handle[0].Upstreams[0].Dial != ":8080" {
		t.Errorf("Expected body Test, got %s", server.Routes[0].Handle[0].Upstreams[0].Dial)
	}
}

func TestConnector_GetCaddyConfigFailsBecauseOfInvalidUrl(t *testing.T) {
	connector := NewConnector("invalid-url")

	_, err := connector.GetCaddyConfig()
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestConnector_GetCaddyConfigFailsBecauseOfEmptyBody(t *testing.T) {
	mockResponse := ""

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(mockResponse))
		} else {
			t.Errorf("Expected %s with method %s, got %s with method %s",
				"/config/", http.MethodGet, r.URL.Path, r.Method)
		}
	}))

	connector := NewConnector(mockServer.URL)
	_, err := connector.GetCaddyConfig()
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestConnector_GetCaddyConfigFailsBecauseOfInvalidJson(t *testing.T) {
	mockResponse := "{"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(mockResponse))
		} else {
			t.Errorf("Expected %s with method %s, got %s with method %s",
				"/config/", http.MethodGet, r.URL.Path, r.Method)
		}
	}))

	connector := NewConnector(mockServer.URL)
	_, err := connector.GetCaddyConfig()
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestConnector_CreateCaddyConfig(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/load" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
		} else {
			t.Errorf("Expected %s with method %s, got %s with method %s",
				"/config/", http.MethodPost, r.URL.Path, r.Method)
		}
	}))

	connector := NewConnector(mockServer.URL)
	err := connector.CreateCaddyConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestConnector_CreateCaddyConfigReturnsError(t *testing.T) {
	connector := NewConnector("invalid-url")

	err := connector.CreateCaddyConfig()
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestConnector_ReplaceRoutes(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/apps/http/servers/srv0/routes/" && r.Method == http.MethodPatch {
			w.WriteHeader(http.StatusOK)
		} else {
			t.Errorf("Expected %s with method %s, got %s with method %s",
				"/config/", http.MethodPost, r.URL.Path, r.Method)
		}
	}))

	connector := NewConnector(mockServer.URL)

	route := NewReverseProxyRoute("subdomain.example.com", 8080)
	routes := []Route{route}

	err := connector.SetRoutes(routes)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestConnector_ReplaceRouteFailsBecauseOfInvalidUrl(t *testing.T) {
	connector := NewConnector("invalid-url")

	route := NewReverseProxyRoute("subdomain.example.com", 8080)
	routes := []Route{route}

	err := connector.SetRoutes(routes)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestNewReverseProxyRoute(t *testing.T) {
	upstreamPort := 8080
	incomingDomain := "subdomain.example.com"
	route := NewReverseProxyRoute(incomingDomain, upstreamPort)

	if len(route.Match) != 1 {
		t.Errorf("Expected 1 route match, got: %d", len(route.Match))
	}
	if len(route.Match[0].Host) != 1 {
		t.Errorf("Expected 1 route match host, got: %d", len(route.Match[0].Host))
	}
	if route.Match[0].Host[0] != incomingDomain {
		t.Errorf("Expected '%s' as domain, got: %s", incomingDomain, route.Match[0].Host[0])
	}
	if len(route.Handle) != 1 {
		t.Errorf("Expected 1 route match handle, got: %d", len(route.Handle))
	}
	if route.Handle[0].Handler != "subroute" {
		t.Errorf("Expected handler subroute, got %s", route.Handle[0].Handler)
	}
	if len(route.Handle[0].Routes) != 1 {
		t.Errorf("Expected 1 route match handle, got %d", len(route.Handle[0].Routes))
	}
	if len(route.Handle[0].Routes[0].Handle) != 1 {
		t.Errorf("Expected 1 route match handle, got %d", len(route.Handle[0].Routes[0].Handle))
	}
	if route.Handle[0].Routes[0].Handle[0].Handler != "reverse_proxy" {
		t.Errorf("Expected handler reverse_proxy, got %s", route.Handle[0].Routes[0].Handle[0].Handler)
	}
	if len(route.Handle[0].Routes[0].Handle[0].Upstreams) != 1 {
		t.Errorf("Expected 1 upstream, got %d", len(route.Handle[0].Routes[0].Handle[0].Upstreams))
	}
	if route.Handle[0].Routes[0].Handle[0].Upstreams[0].Dial != ":"+strconv.Itoa(upstreamPort) {
		t.Errorf("Expected handler reverse_proxy, got %s", route.Handle[0].Routes[0].Handle[0].Upstreams[0].Dial)
	}
}
