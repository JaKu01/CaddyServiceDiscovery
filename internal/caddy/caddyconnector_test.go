package caddy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnector_GetCaddyConfig(t *testing.T) {
	mockResponse := "{\"apps\":{\"http\":{\"servers\":{\"exampleServer\":{\"listen\":[\":443\"],\"routes\":[{\"handle\":[{\"handler\":\"static_response\",\"body\":\"Test\"}]}]}}}}}\n"

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
	if server.Routes[0].Handle[0].Handler != "static_response" {
		t.Errorf("Expected handler static_response, got %s", server.Routes[0].Handle[0].Handler)
	}
	if server.Routes[0].Handle[0].Body != "Test" {
		t.Errorf("Expected body Test, got %s", server.Routes[0].Handle[0].Body)
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

func TestConnector_ReplaceServers(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/apps/http/servers" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
		} else {
			t.Errorf("Expected %s with method %s, got %s with method %s",
				"/config/", http.MethodPost, r.URL.Path, r.Method)
		}
	}))

	connector := NewConnector(mockServer.URL)

	server := NewReverseProxyServer(8080, "127.0.0.1:9000")
	serverMap := make(map[string]Server)
	serverMap["exampleServer"] = server

	err := connector.SetServers(serverMap)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestConnector_ReplaceServersFailsBecauseOfInvalidUrl(t *testing.T) {
	connector := NewConnector("invalid-url")

	server := NewReverseProxyServer(8080, "127.0.0.1:9000")
	serverMap := make(map[string]Server)
	serverMap["exampleServer"] = server

	err := connector.SetServers(serverMap)
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestNewReverseProxyServer(t *testing.T) {
	port := 8080
	upstream := "127.0.0.1:9000"
	server := NewReverseProxyServer(port, upstream)

	if len(server.Listen) != 1 || server.Listen[0] != ":8080" {
		t.Errorf("Expected listen port ':8080', got: %v", server.Listen)
	}
	if len(server.Routes) != 1 {
		t.Errorf("Expected 1 route, got: %d", len(server.Routes))
	}
	if len(server.Routes[0].Handle) != 1 {
		t.Errorf("Expected 1 handler, got: %d", len(server.Routes[0].Handle))
	}
	handler := server.Routes[0].Handle[0]
	if handler.Handler != "reverse_proxy" {
		t.Errorf("Expected handler 'reverse_proxy', got: %s", handler.Handler)
	}
	if len(handler.Upstreams) != 1 || handler.Upstreams[0].Dial != upstream {
		t.Errorf("Expected upstream '%s', got: %v", upstream, handler.Upstreams)
	}
}
