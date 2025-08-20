package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"net/url"
)

import (
	mslogger "mockserver/logger"
	msUtils "mockserver/utils"
)

// Route validation regex (path must start with / and contain only valid chars)
var validPathRegex = regexp.MustCompile(`^\/[a-zA-Z0-9\/\-_{}]*$`)

// [IMP_FUNC]
func validateAndApplyDefaults(cfg *Config, configFilePath string) error {

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 5000
		mslogger.LogWarn("Config: server.port not set → using default 5000")
	}

	if cfg.Server.APIPrefix == "" {
		cfg.Server.APIPrefix = ""
		// [OPTIONAL_LOG] mslogger.LogWarn("Config: server.api_prefix not set → using default '/'")
	}

	if cfg.Server.DefaultHeaders == nil {
		cfg.Server.DefaultHeaders = map[string]string{"Content-Type": "application/json"}
		// [OPTIONAL_LOG] mslogger.LogWarn("Config: server.default_headers not set → using default 'Content-Type: application/json'")
	}

	if cfg.Server.SwaggerUIPath == "" {
		cfg.Server.SwaggerUIPath = "/docs"
		// [OPTIONAL_LOG] mslogger.LogInfo("Config: server.swagger_ui_path not set → using default '/docs'")
	}

	// Auth validation
	if cfg.Server.Auth != nil && cfg.Server.Auth.Enabled {
		if err := validateAuth(cfg.Server.Auth); err != nil {
			return err
		}
	}

	// Routes validation
	for i, route := range cfg.Routes {
		if err := validateRoute(&route, configFilePath); err != nil {
			return fmt.Errorf("route[%d] '%s' validation failed: %w", i, route.Name, err)
		}
		cfg.Routes[i] = route
	}

	return nil
}

func validateAuth(auth *AuthConfig) error {
	if auth.Type == "" {
		return fmt.Errorf("auth.type is required when auth.enabled = true")
	}
	if auth.In != "header" && auth.In != "query" {
		return fmt.Errorf("auth.in must be either 'header' or 'query'")
	}
	return nil
}

func validateRoute(route *RouteConfig, configFilePath string) error {

	// Method validation
	if _, ok := msUtils.AllowedMethods[strings.ToUpper(route.Method)]; !ok {
		return fmt.Errorf("invalid method '%s'", route.Method)
	}

	// Path validation
	if !validPathRegex.MatchString(route.Path) {
		return fmt.Errorf("invalid path '%s': must start with '/' and contain only letters, numbers, '-', '_', '{', '}'", route.Path)
	}

	// Fetch validation
	if route.Fetch != nil {
		if err := validateFetch(route.Fetch, route.Path); err != nil {
			return err
		}
	}

	// Mock validation
	if route.Mock != nil {
		if err := validateMock(route.Mock, route.Path, configFilePath); err != nil {
			return err
		}
	}

	return nil
}

func validateFetch(fetch *FetchConfig, routePath string) error {
	if fetch.URL == "" {
		return fmt.Errorf("[Route %s] fetch.url is required", routePath)
	}

	parsed, err := url.Parse(fetch.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("[Route %s] fetch.url is invalid: '%s'", routePath, fetch.URL)
	}

	return nil
}

func validateMock(mock *MockConfig, routePath string, configFilePath string) error {
	if !strings.HasSuffix(mock.File, ".json") {
		return fmt.Errorf("[Route %s] mock.file must be a .json file, got '%s'", routePath, mock.File)
	}

	 mockFilePath := msUtils.ResolveMockFilePath(configFilePath, mock.File)

	if _, err := os.Stat(mockFilePath); err != nil {
		return fmt.Errorf("[Route %s] mock.file not found: '%s'", routePath, mock.File)
	}

	if mock.Status != 0 {
		if mock.Status < 100 || mock.Status > 599 {
			return fmt.Errorf("[Route %s] mock.status must be between 100 and 599, got %d", routePath, mock.Status)
		}
	}

	if mock.DelayMs < 0 {
		return fmt.Errorf("[Route %s] mock.delay_ms cannot be negative, got %d", routePath, mock.DelayMs)
	}

	return nil
}
