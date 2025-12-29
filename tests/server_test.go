package tests

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mockserver/config"
	"mockserver/server"
)

// Dummy embed FS
//
//go:embed server_test.go
var testEmbedFS embed.FS


func makeRequest(method, url string, body interface{}, headers map[string]string) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}


func createSafeConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:      5000,
			APIPrefix: "/v1",
			Debug:   &config.DebugConfig{Enabled: false, Path: "/__debug"},
			Console: &config.ConsoleConfig{Enabled: false, Path: "/console", Auth: &config.ConsoleAuthConfig{Enabled: true}},
			CORS:    &config.CORSConfig{Enabled: false},
			Auth:    &config.AuthConfig{Enabled: false},
		},
		Routes: []config.RouteConfig{},
	}
}


// 1. SIMPLE MOCK TEST
func TestIntegration_SimpleMock(t *testing.T) {
	cfg := createSafeConfig()

	cfg.Routes = []config.RouteConfig{
		{
			Name:   "Test Route",
			Method: "GET",
			Path:   "/hello",
			Mock: &config.MockConfig{
				Status: 200,
				Body:   map[string]interface{}{"message": "world"},
			},
		},
	}

	app := server.StartServer(cfg, "", testEmbedFS)

	// Send a request
	req := makeRequest("GET", "/v1/hello", nil, nil)
	resp, err := app.Test(req, -1)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"message": "world"}`, string(bodyBytes))
}


// 2. LOGIC ENGINE & CASES TEST
func TestIntegration_LogicCases(t *testing.T) {
	cfg := createSafeConfig()
	cfg.Server.APIPrefix = "/api"

	cfg.Routes = []config.RouteConfig{
		{
			Name:   "Dynamic Pricing",
			Method: "POST",
			Path:   "/price",
			Cases: []config.CaseConfig{
				{
					When: "request.body.type == 'vip'",
					Then: config.CResponse{
						Status: 200,
						Body:   map[string]interface{}{"price": 50},
					},
				},
			},
			Mock: &config.MockConfig{ // Default Case
				Status: 200,
				Body:   map[string]interface{}{"price": 100},
			},
		},
	}

	app := server.StartServer(cfg, "", testEmbedFS)

	// Senaryo A: VIP
	reqVIP := makeRequest("POST", "/api/price", map[string]string{"type": "vip"}, nil)
	respVIP, _ := app.Test(reqVIP)
	bodyVIP, _ := io.ReadAll(respVIP.Body)
	assert.Equal(t, 200, respVIP.StatusCode)
	assert.JSONEq(t, `{"price": 50}`, string(bodyVIP))

	// Senaryo B: Normal
	reqNorm := makeRequest("POST", "/api/price", map[string]string{"type": "normal"}, nil)
	respNorm, _ := app.Test(reqNorm)
	bodyNorm, _ := io.ReadAll(respNorm.Body)
	assert.JSONEq(t, `{"price": 100}`, string(bodyNorm))
}


// 3. STATEFUL INTEGRATION TEST
func TestIntegration_StatefulFlow(t *testing.T) {
	cfg := createSafeConfig()

	cfg.Routes = []config.RouteConfig{
		{
			Name:     "Create User",
			Method:   "POST",
			Path:     "/users",
			Stateful: &config.StatefulConfig{Collection: "users", Action: "create", IDField: "id"},
			Mock: &config.MockConfig{
				Status: 200,
				Body:   "{{state.created}}",
			},
			BodySchema: &config.JSONSchema{
				Type: "object",
				Properties: map[string]*config.JSONSchema{
					"id":    {Type: "integer"},
					"name":  {Type: "string"},
					"email": {Type: "string"},
					"role":  {Type: "string"},
				},
			},
		},
		{
			Name:     "Get User",
			Method:   "GET",
			Path:     "/users/{id}",
			Stateful: &config.StatefulConfig{Collection: "users", Action: "get", IDField: "id"},

			Mock: &config.MockConfig{
				Status: 200,
				Body:   "{{state.item}}",
			},

			BodySchema: &config.JSONSchema{
				Type: "object",
				Properties: map[string]*config.JSONSchema{
					"id":    {Type: "integer"},
					"name":  {Type: "string"},
					"email": {Type: "string"},
					"role":  {Type: "string"},
				},
			},
		},
	}

	app := server.StartServer(cfg, "", testEmbedFS)

	// Step 1: Create User
	newUser := map[string]interface{}{"id": 123, "name": "CTO"}
	reqCreate := makeRequest("POST", "/v1/users", newUser, nil)
	respCreate, _ := app.Test(reqCreate)
	assert.Equal(t, 200, respCreate.StatusCode)

	// Step 2: Get User
	reqGet := makeRequest("GET", "/v1/users/123", nil, nil)
	respGet, _ := app.Test(reqGet)

	bodyGet, _ := io.ReadAll(respGet.Body)
	assert.Equal(t, 200, respGet.StatusCode)
	assert.Contains(t, string(bodyGet), "CTO")
}


// 4. AUTH TEST
func TestIntegration_Auth(t *testing.T) {
	cfg := createSafeConfig()
	cfg.Server.APIPrefix = "/secure"

	cfg.Server.Auth = &config.AuthConfig{
		Enabled: true,
		Type:    "apiKey",
		In:      "header",
		Name:    "X-Secret",
		Keys:    []string{"super-secret-key"},
	}

	cfg.Routes = []config.RouteConfig{
		{
			Name:   "Secret Data",
			Method: "GET",
			Path:   "/data",
			Mock:   &config.MockConfig{Status: 200, Body: "Success"},
		},
	}

	app := server.StartServer(cfg, "", testEmbedFS)

	// Scenario 1: Keyless (Fail)
	reqFail := makeRequest("GET", "/secure/data", nil, nil)
	respFail, _ := app.Test(reqFail)
	assert.Equal(t, 401, respFail.StatusCode)

	// Scenario 2: The Right Key (Success)
	reqSuccess := makeRequest("GET", "/secure/data", nil, map[string]string{"X-Secret": "super-secret-key"})
	respSuccess, _ := app.Test(reqSuccess)
	assert.Equal(t, 200, respSuccess.StatusCode)
}


// 5. FETCH (PROXY) TEST
func TestIntegration_Fetch(t *testing.T) {
	cfg := createSafeConfig()
	cfg.Server.APIPrefix = "/proxy"

	cfg.Routes = []config.RouteConfig{
		{
			Name:   "Google Proxy",
			Method: "GET",
			Path:   "/google",
			Fetch:  &config.FetchConfig{URL: "https://www.google.com"},
		},
	}

	app := server.StartServer(cfg, "", testEmbedFS)

	req := makeRequest("GET", "/proxy/google", nil, nil)
	resp, _ := app.Test(req, 5000)

	assert.Equal(t, 200, resp.StatusCode)
}
