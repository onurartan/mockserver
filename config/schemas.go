package config

import (
	mslogger "mockserver/logger"
)

type ParamDef struct {
	// Data type of the parameter (string, number, boolean, etc.)
	Type string `json:"type" yaml:"type"`

	// Human-readable explanation of the parameter
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Whether this parameter is mandatory
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`

	// Allowed values if restricted to a fixed set
	Enum []string `json:"enum,omitempty" yaml:"enum,omitempty"`

	// Example value for documentation or testing
	Example interface{} `json:"example,omitempty" yaml:"example,omitempty"`
}

// GroupConfig groups multiple routes under a category
type GroupConfig struct {
	// Unique group name (used in docs/Swagger tags)
	Name string `json:"name" yaml:"name"`

	// Optional description of the group
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type CORSConfig struct {
	// Enable or disable CORS globally
	Enabled bool `json:"enabled" yaml:"enabled"`

	// List of allowed origins (e.g., ["*"] for all)
	AllowOrigins []string `json:"allow_origins" yaml:"allow_origins"`

	// Allowed HTTP methods (e.g., GET, POST)
	AllowMethods []string `json:"allow_methods" yaml:"allow_methods"`

	// Allowed request headers
	AllowHeaders []string `json:"allow_headers" yaml:"allow_headers"`

	// Allow cookies/auth headers across origins
	AllowCredentials bool `json:"allow_credentials" yaml:"allow_credentials"`
}

type AuthConfig struct {
	// Enable or disable authentication
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Authentication type: "apikey" or "bearer"
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// Where to pass the key: "header" or "query"
	In string `json:"in,omitempty" yaml:"in,omitempty"`

	// Parameter name (e.g., "Authorization", "X-API-Key")
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// List of valid API keys or tokens
	Keys []string `json:"keys,omitempty" yaml:"keys,omitempty"`
}

type DebugConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Path    string `json:"path" yaml:"path"`
}

type ConsoleAuthConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type ConsoleConfig struct {
	Enabled bool               `json:"enabled" yaml:"enabled"`
	Path    string             `json:"path" yaml:"path"`
	Auth    *ConsoleAuthConfig `json:"auth" yaml:"auth"`
}

type ServerConfig struct {
	// Port on which the server will run
	Port int `json:"port" yaml:"port"`

	Console *ConsoleConfig `json:"console" yaml:"console"`

	Debug *DebugConfig `json:"debug,omitempty" yaml:"debug,omitempty"`

	// Global prefix for all API routes (e.g., "/v1")
	APIPrefix string `json:"api_prefix" yaml:"api_prefix"`

	// Headers applied to every response by default
	DefaultHeaders map[string]string `json:"default_headers" yaml:"default_headers"`

	// Global response delay (in milliseconds)
	DefaultDelayMs int `json:"default_delay_ms" yaml:"default_delay_ms"`

	// Path to expose Swagger UI (e.g., "/docs")
	SwaggerUIPath string `json:"swagger_ui_path" yaml:"swagger_ui_path"`

	// CORS configuration
	CORS *CORSConfig `json:"cors" yaml:"cors"`

	// Global authentication configuration
	Auth *AuthConfig `json:"auth,omitempty" yaml:"auth,omitempty"`
}

// JSONSchema: Represents a standard JSON Schema (Draft 7 compatible).
// Supports recursive structures for nested objects and arrays.
type JSONSchema struct {
	// Data type (e.g., "string", "integer", "object", "array")
	Type string `yaml:"type,omitempty" json:"type,omitempty"`

	// Human-readable description of the field's purpose
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// List of mandatory property names (only for "object" type)
	Required []string `yaml:"required,omitempty" json:"required,omitempty"`

	// Key-value definitions for object fields (recursive)
	Properties map[string]*JSONSchema `yaml:"properties,omitempty" json:"properties,omitempty"`

	// Schema for array elements (only for "array" type)
	Items *JSONSchema `yaml:"items,omitempty" json:"items,omitempty"`

	// List of strictly allowed values
	Enum []interface{} `yaml:"enum,omitempty" json:"enum,omitempty"`

	// Minimum numeric value allowed (inclusive)
	Minimum *float64 `yaml:"minimum,omitempty" json:"minimum,omitempty"`

	// Maximum numeric value allowed (inclusive)
	Maximum *float64 `yaml:"maximum,omitempty" json:"maximum,omitempty"`

	// Minimum character count for strings
	MinLength *int `yaml:"minLength,omitempty" json:"minLength,omitempty"`

	// Maximum character count for strings
	MaxLength *int `yaml:"maxLength,omitempty" json:"maxLength,omitempty"`

	// Regular expression pattern for string validation
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty"`

	// If true, allows keys not defined in 'Properties'
	AdditionalProperties bool `yaml:"additional_properties,omitempty" json:"additionalProperties,omitempty"`
}

type CResponse struct {
	// HTTP status code
	Status int `json:"status" yaml:"status"`

	// Response body
	Body interface{} `json:"body,omitempty" yaml:"body,omitempty"`

	// Response headers
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Response delay (in milliseconds)
	DelayMs int `json:"delay_ms,omitempty" yaml:"delay_ms,omitempty"`
}

type StatefulConfig struct {
	Collection string `json:"collection" yaml:"collection"`
	Action     string `json:"action" yaml:"action"` // create|get|update|delete|list
	IDField    string `json:"id_field" yaml:"id_field"`
}

type CaseConfig struct {
	// Boolean expression to evaluate
	When string `json:"when" yaml:"when"`

	// Response to return if condition matches
	Then CResponse `json:"then" yaml:"then"`
}

type MockConfig struct {
	// Inline mock response body (main uses)
	Body interface{} `json:"body,omitempty" yaml:"body,omitempty"`

	// Optional JSON file path (legacy / stateless)
	File string `json:"file,omitempty" yaml:"file,omitempty"`

	// HTTP status code
	Status int `json:"status" yaml:"status"`

	// Custom headers
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Artificial delay
	DelayMs int `json:"delay_ms,omitempty" yaml:"delay_ms,omitempty"`
}

type FetchConfig struct {
	// Target URL to fetch or proxy data from
	URL string `json:"url" yaml:"url"`

	// HTTP method for the fetch (default: GET)
	Method string `json:"method,omitempty" yaml:"method,omitempty"`

	// Additional headers to forward with the request
	Headers map[string]string `json:"headers" yaml:"headers"`

	// Extra query params to append to the request
	QueryParams map[string]string `json:"query_params" yaml:"query_params"`

	// If true, pass through upstream HTTP status
	PassStatus bool `json:"pass_status" yaml:"pass_status"`

	// Artificial delay before returning fetch response
	DelayMs int `json:"delay_ms" yaml:"delay_ms"`

	// Timeout for the external request
	TimeoutMs int `json:"timeout_ms,omitempty" yaml:"timeout_ms,omitempty"`
}

type RouteConfig struct {
	// Unique name of the route
	Name string `json:"name" yaml:"name"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Tag used for grouping in Swagger/docs
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty"`

	// HTTP method (GET, POST, etc.)
	Method string `json:"method" yaml:"method"`

	// Endpoint path (supports params like /users/:id)
	Path string `json:"path" yaml:"path"`

	// Default status code if mock/fetch not used
	Status int `json:"status,omitempty" yaml:"status,omitempty"`

	// Custom response headers for this route
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Response delay specific to this route
	DelayMs int `json:"delay_ms,omitempty" yaml:"delay_ms,omitempty"`

	// Path parameters definition
	PathParams map[string]ParamDef `json:"path_params,omitempty" yaml:"path_params,omitempty"`

	// Query parameters definition
	Query map[string]ParamDef `json:"query,omitempty" yaml:"query,omitempty"`

	// Expected request headers definition
	RequestHeaders map[string]ParamDef `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`

	// Schema for request body validation
	// BodySchema map[string]interface{} `json:"body_schema,omitempty" yaml:"body_schema,omitempty"` [OLD]
	BodySchema *JSONSchema `json:"body_schema,omitempty" yaml:"body_schema,omitempty"`

	// Example body for documentation/testing
	BodyExample interface{} `json:"body_example,omitempty" yaml:"body_example,omitempty"`

	// Static mock response configuration
	Mock *MockConfig `json:"mock,omitempty" yaml:"mock,omitempty"`

	// Proxy/fetch response configuration
	Fetch *FetchConfig `json:"fetch,omitempty" yaml:"fetch,omitempty"`

	// Conditional responses (rule-based behavior)
	Cases    []CaseConfig    `json:"cases,omitempty" yaml:"cases,omitempty"`
	Stateful *StatefulConfig `json:"stateful,omitempty" yaml:"stateful,omitempty"`

	Default *CResponse `json:"default,omitempty" yaml:"default,omitempty"`

	// Route-specific authentication override
	Auth *AuthConfig `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type Config struct {
	// Optional JSON schema reference for validation
	Schema string `json:"$schema,omitempty" yaml:"$schema,omitempty"`

	// Global server configuration
	Server ServerConfig `json:"server" yaml:"server"`

	// Optional groups to organize routes
	Groups []GroupConfig `json:"groups,omitempty" yaml:"groups,omitempty"`

	// List of all API routes
	Routes []RouteConfig `json:"routes" yaml:"routes"`
}

// helpers
func (s *ServerConfig) ApplyServerDefaults() {

	
	if s.Port == 0 {
		s.Port = 5000
		mslogger.LogWarn("Config: server.port not set → using default 5000")
	}

	if s.APIPrefix == "" {
		s.APIPrefix = ""
		// [OPTIONAL_LOG] mslogger.LogWarn("Config: server.api_prefix not set → using default '/'")
	}

	if s.DefaultDelayMs == 0 {
		s.DefaultDelayMs = 0
		// [OPTIONAL_LOG] mslogger.LogWarn("Config: server.default_delay_ms not set → using default 0")
	}

	if s.DefaultHeaders == nil {
		s.DefaultHeaders = map[string]string{"Content-Type": "application/json"}
		// [OPTIONAL_LOG] mslogger.LogWarn("Config: server.default_headers not set → using default 'Content-Type: application/json'")
	}

	if s.SwaggerUIPath == "" {
		s.SwaggerUIPath = "/docs"
		// [OPTIONAL_LOG] mslogger.LogInfo("Config: server.swagger_ui_path not set → using default '/docs'")
	}

	// --- Debug ---
	if s.Debug == nil {
		s.Debug = &DebugConfig{}
	}
	if s.Debug.Path == "" {
		s.Debug.Path = "/__debug"
	}

	if s.Console == nil {
		s.Console = &ConsoleConfig{
			Enabled: true,
		}
	}

	if s.Console.Path == "" {
		s.Console.Path = "/console"
	}

	if s.Console.Auth == nil {

		s.Console.Auth = &ConsoleAuthConfig{
			Enabled:  true,
			Username: "admin",
			Password: "123",
		}

		mslogger.LogWarn("Console auth default credentials are in use (admin/1**)")

	}

	if s.CORS == nil {
		s.CORS = &CORSConfig{}
	}

	if s.CORS.Enabled {
		if len(s.CORS.AllowOrigins) == 0 {
			s.CORS.AllowOrigins = []string{"*"}
		}
		if len(s.CORS.AllowMethods) == 0 {
			s.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
		}
		if len(s.CORS.AllowHeaders) == 0 {
			s.CORS.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
		}
	}

}
