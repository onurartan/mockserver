package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTempFile creates a temporary file with the given content for testing isolation.
func createTempFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "Geçici dosya oluşturulamadı")
	return path
}

// TestValidateAndApplyDefaults verifies that the system applies safe default values
// (e.g., Port 5000, JSON headers) when the configuration is empty.
func TestValidateAndApplyDefaults(t *testing.T) {
	t.Run("Should apply default values to empty config", func(t *testing.T) {
		cfg := &Config{}

		validateAndApplyDefaults(cfg, "")

		assert.Equal(t, 5000, cfg.Server.Port)
		assert.Equal(t, "", cfg.Server.APIPrefix)
	
		assert.NotNil(t, cfg.Server.Debug)
		assert.Equal(t, "/__debug", cfg.Server.Debug.Path)

		assert.NotNil(t, cfg.Server.Console)
		assert.Equal(t, "/console", cfg.Server.Console.Path)
		
		assert.NotNil(t, cfg.Server.SwaggerUIPath)
		assert.Equal(t, "/docs", cfg.Server.SwaggerUIPath)
		
		assert.NotNil(t, cfg.Server.CORS)
	})

	t.Run("Should respect existing values", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Port: 8080,
				Debug: &DebugConfig{
					Enabled: true,
					Path:    "/custom-debug",
				},
			},
		}

		validateAndApplyDefaults(cfg, "")

		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "/custom-debug", cfg.Server.Debug.Path)
		assert.NotNil(t, cfg.Server.Console)
	})
	
	t.Run("Should set default CORS values if enabled", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				CORS: &CORSConfig{
					Enabled: true,
				},
			},
		}
		
		validateAndApplyDefaults(cfg, "")
		
		assert.NotEmpty(t, cfg.Server.CORS.AllowOrigins)
		assert.Equal(t, "*", cfg.Server.CORS.AllowOrigins[0])
	})
}


// TestLoadConfig_ComplexIntegration simulates a comprehensive production scenario.
// It verifies the mapping of Logic Cases, Stateful definitions, and Auth settings
// from YAML to the Go struct.
func TestLoadConfig_ComplexIntegration(t *testing.T) {
	yamlContent := `
server:
  port: 8080
  api_prefix: "/api/v1"
  auth:
    enabled: true
    type: "apiKey"
    in: "query"
    name: "apiKey"
    keys:
      - "secret"
      - "team-secret"

routes:
  - name: "Create Order"
    method: "POST"
    path: "/orders"
    body_schema:
      type: "object"
      properties:
        id: { type: "string" }
      required:
        - "id"
    stateful:
      collection: "orders"
      action: "create"
      id_field: "order_id"

    mock:
      status: 201
      body: { "id": "{{request.body.id}}" }

  - name: "Dynamic Payment"
    method: "POST"
    path: "/payment"
    cases:
      - when: "request.body.amount > 1000"
        then:
          status: 400
          body: { "error": "Over limit" }
          delay_ms: 100
      - when: "request.body.currency == 'USD'"
        then:
          status: 200
          body: { "status": "paid" }
    default:
      status: 500

`
	tmpDir := t.TempDir()
	configFile := createTempFile(t, tmpDir, "mockserver_complex.yaml", yamlContent)

	cfg, err := LoadConfig(configFile)

	// Critical Integrity Checks
	require.NoError(t, err, "Error occurred while loading complex config")
	require.NotNil(t, cfg)

	// Verify Server Settings
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "/api/v1", cfg.Server.APIPrefix)
	assert.True(t, cfg.Server.Auth.Enabled)

	// Verify Stateful Route (Index 0)
	statefulRoute := cfg.Routes[0]
	assert.Equal(t, "Create Order", statefulRoute.Name)
	assert.Equal(t, "secret", cfg.Server.Auth.Keys[0])
	assert.Equal(t, "team-secret", cfg.Server.Auth.Keys[1])
	require.NotNil(t, statefulRoute.Stateful, "Stateful configuration not read")
	assert.Equal(t, "orders", statefulRoute.Stateful.Collection)
	assert.Equal(t, "create", statefulRoute.Stateful.Action)
	assert.Equal(t, "order_id", statefulRoute.Stateful.IDField)

	// Verify Logic Cases Route (Index 1)
	logicRoute := cfg.Routes[1]
	require.Len(t, logicRoute.Cases, 2, "The number of cases was misread.")

	assert.Equal(t, "request.body.amount > 1000", logicRoute.Cases[0].When)
	assert.Equal(t, 400, logicRoute.Cases[0].Then.Status)
	assert.Equal(t, 100, logicRoute.Cases[0].Then.DelayMs)

	assert.Equal(t, "request.body.currency == 'USD'", logicRoute.Cases[1].When)
	assert.Equal(t, 200, logicRoute.Cases[1].Then.Status)
}

// TestRouteValidation_Rules uses table-driven tests to verify various user errors
// such as missing files, invalid extensions, or missing required fields.
func TestRouteValidation_Rules(t *testing.T) {
	tmpDir := t.TempDir()
	validJson := createTempFile(t, tmpDir, "data.json", "{}")
	invalidExt := createTempFile(t, tmpDir, "data.txt", "dummy")

	tests := []struct {
		name        string
		route       RouteConfig
		mockConfig  *MockConfig
		fetchConfig *FetchConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid Mock File",
			mockConfig: &MockConfig{
				File: validJson,
			},
			expectError: false,
		},
		{
			name: "Missing Mock File",
			mockConfig: &MockConfig{
				File: "non_existent_file.json",
			},
			expectError: true,
			errorMsg:    "file not found",
		},
		{
			name: "Invalid File Extension",
			mockConfig: &MockConfig{
				File: invalidExt,
			},
			expectError: true,
			errorMsg:    "unsupported mock file extension",
		},
		{
			name: "Fetch without URL",
			fetchConfig: &FetchConfig{
				URL: "",
			},
			expectError: true,
			errorMsg:    "fetch url is required",
		},
		{
			name: "Valid Fetch",
			fetchConfig: &FetchConfig{
				URL: "https://api.example.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			if tt.mockConfig != nil {
				err = validateMock(tt.mockConfig, "/test", "")
			}

			if tt.fetchConfig != nil {
				err = validateFetch(tt.fetchConfig, "/test")
			}

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateStateful_RequiredFields ensures that stateful routes strictly enforce
// the presence of 'collection' and 'action' fields.
func TestValidateStateful_RequiredFields(t *testing.T) {
	// Case 1: Missing Collection
	badConfig1 := &StatefulConfig{
		Action: "create",
	}
	err1 := validateStateful(badConfig1, "/users")
	assert.Error(t, err1, "If the collection is missing, it should throw an error.")

	// Case 2: Missing Action
	badConfig2 := &StatefulConfig{
		Collection: "users",
	}
	err2 := validateStateful(badConfig2, "/users")
	assert.Error(t, err2, "If there is no action, it should return an error.")

	// Case 3: Valid Config
	goodConfig := &StatefulConfig{
		Collection: "users",
		Action:     "create",
	}
	err3 := validateStateful(goodConfig, "/users")
	assert.NoError(t, err3, "If the configuration is correct, it should not give an error.")
}


