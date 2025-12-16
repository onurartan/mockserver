package config

type ParamDef struct {
	// Data type of the parameter (string, number, boolean, etc.)
	Type string `json:"type"`

	// Human-readable explanation of the parameter
	Description string `json:"description,omitempty"`

	// Whether this parameter is mandatory
	Required bool `json:"required,omitempty"`

	// Allowed values if restricted to a fixed set
	Enum []string `json:"enum,omitempty"`

	// Example value for documentation or testing
	Example interface{} `json:"example,omitempty"`
}

// GroupConfig groups multiple routes under a category
type GroupConfig struct {
	// Unique group name (used in docs/Swagger tags)
	Name string `json:"name"`

	// Optional description of the group
	Description string `json:"description,omitempty"`
}

type CORSConfig struct {
	// Enable or disable CORS globally
	Enabled bool `json:"enabled"`

	// List of allowed origins (e.g., ["*"] for all)
	AllowOrigins []string `json:"allow_origins"`

	// Allowed HTTP methods (e.g., GET, POST)
	AllowMethods []string `json:"allow_methods"`

	// Allowed request headers
	AllowHeaders []string `json:"allow_headers"`

	// Allow cookies/auth headers across origins
	AllowCredentials bool `json:"allow_credentials"`
}

type AuthConfig struct {
	// Enable or disable authentication
	Enabled bool `json:"enabled"`

	// Authentication type: "apikey" or "bearer"
	Type string `json:"type,omitempty"`

	// Where to pass the key: "header" or "query"
	In string `json:"in,omitempty"`

	// Parameter name (e.g., "Authorization", "X-API-Key")
	Name string `json:"name,omitempty"`

	// List of valid API keys or tokens
	Keys []string `json:"keys,omitempty"`
}

type DebugConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

type ServerConfig struct {
	// Port on which the server will run
	Port int `json:"port"`
	
	Debug *DebugConfig `json:"debug,omitempty"`

	// Global prefix for all API routes (e.g., "/v1")
	APIPrefix string `json:"api_prefix"`

	// Headers applied to every response by default
	DefaultHeaders map[string]string `json:"default_headers"`

	// Global response delay (in milliseconds)
	DefaultDelayMs int `json:"default_delay_ms"`

	// Path to expose Swagger UI (e.g., "/docs")
	SwaggerUIPath string `json:"swagger_ui_path"`

	// CORS configuration
	CORS CORSConfig `json:"cors"`

	// Global authentication configuration
	Auth *AuthConfig `json:"auth,omitempty"`
}

type MockConfig struct {
	// Path to mock JSON file used as response
	File string `json:"file"`

	// HTTP status code for the mock response
	Status int `json:"status"`

	// Custom headers for the mock response
	Headers map[string]string `json:"headers"`

	// Artificial response delay (in milliseconds)
	DelayMs int `json:"delay_ms"`
}

type FetchConfig struct {
	// Target URL to fetch or proxy data from
	URL string `json:"url"`

	// HTTP method for the fetch (default: GET)
	Method string `json:"method,omitempty"`

	// Additional headers to forward with the request
	Headers map[string]string `json:"headers"`

	// Extra query params to append to the request
	QueryParams map[string]string `json:"query_params"`

	// If true, pass through upstream HTTP status
	PassStatus bool `json:"pass_status"`

	// Artificial delay before returning fetch response
	DelayMs int `json:"delay_ms"`

	// Timeout for the external request
	TimeoutMs int `json:"timeout_ms,omitempty"`
}

type RouteConfig struct {
	// Unique name of the route
	Name string `json:"name"`

	Description string `json:"description,omitempty"`
	
	// Tag used for grouping in Swagger/docs
	Tag string `json:"tag,omitempty"`

	// HTTP method (GET, POST, etc.)
	Method string `json:"method"`

	// Endpoint path (supports params like /users/:id)
	Path string `json:"path"`

	// Default status code if mock/fetch not used
	Status int `json:"status,omitempty"`

	// Custom response headers for this route
	Headers map[string]string `json:"headers,omitempty"`

	// Response delay specific to this route
	DelayMs int `json:"delay_ms,omitempty"`

	// Path parameters definition
	PathParams map[string]ParamDef `json:"path_params,omitempty"`

	// Query parameters definition
	Query map[string]ParamDef `json:"query,omitempty"`

	// Expected request headers definition
	RequestHeaders map[string]ParamDef `json:"request_headers,omitempty"`

	// Schema for request body validation
	BodySchema map[string]interface{} `json:"body_schema,omitempty"`

	// Example body for documentation/testing
	BodyExample interface{} `json:"body_example,omitempty"`

	// Static mock response configuration
	Mock *MockConfig `json:"mock,omitempty"`

	// Proxy/fetch response configuration
	Fetch *FetchConfig `json:"fetch,omitempty"`

	// Route-specific authentication override
	Auth *AuthConfig `json:"auth,omitempty"`
}

type Config struct {
	// Optional JSON schema reference for validation
	Schema string `json:"$schema,omitempty"`

	// Global server configuration
	Server ServerConfig `json:"server"`

	// Optional groups to organize routes
	Groups []GroupConfig `json:"groups,omitempty"`

	// List of all API routes
	Routes []RouteConfig `json:"routes"`
}
