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

// Cases Conf
const maxCasesPerRoute = 20

var rootRegex = regexp.MustCompile(
	`(request\.)?(body|query|headers|path)\.[a-zA-Z0-9_]+|method\b`,
)
var allowedConditionRoots = []string{
	"body.",
	"query.",
	"headers.",
	"path.",
	"method",
}

// [IMP_FUNC]
func validateAndApplyDefaults(cfg *Config, configFilePath string) error {

	cfg.Server.ApplyServerDefaults()

	// Auth validation
	if cfg.Server.Auth != nil && cfg.Server.Auth.Enabled {
		if err := validateAuth(cfg.Server.Auth); err != nil {
			return err
		}
	}

	if cfg.Server.Debug != nil {
		if !validPathRegex.MatchString(cfg.Server.Debug.Path) {
			return fmt.Errorf("invalid debug path '%s': must start with '/' ...", cfg.Server.Debug.Path)
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

	// Stateful Validation
	if route.Stateful != nil {

		isWriteAction := route.Stateful.Action == "create" || route.Stateful.Action == "update"
		if route.BodySchema == nil && isWriteAction {
			return fmt.Errorf("stateful route '%s' requires 'body_schema' for data integrity", route.Path)
		}

		if len(route.Cases) == 0 && route.Mock == nil {
			return fmt.Errorf("stateful route '%s' must define a 'mock' response or 'cases' to return the state", route.Path)
		}

		if route.Fetch != nil {
			mslogger.LogWarn(fmt.Sprintf("Route '%s': both stateful and fetch defined. Stateful logic will run before proxying.", route.Path))
		}
	}

	if err := validateStateful(route.Stateful, route.Path); err != nil {
		return err
	}

	// Cases validation
	if len(route.Cases) > 0 {
		if err := validateCases(route.Cases, route.Path); err != nil {
			return err
		}
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

		if route.Stateful != nil && len(route.Cases) == 0 {
			if route.Mock.Body == nil && route.Mock.File == "" {
				return fmt.Errorf("stateful route '%s' requires a mock body (e.g. '{{state.created}}') to return results", route.Path)
			}
		}
	}

	if len(route.Cases) > 0 && route.Mock != nil {
		mslogger.LogWarn(
			fmt.Sprintf("Route '%s': cases defined, mock will be used only if no case matches", route.Path),
		)
	}

	if len(route.Cases) > 0 && route.Fetch != nil {
		mslogger.LogWarn(
			fmt.Sprintf("Route '%s': cases defined, fetch will be used only if no case matches", route.Path),
		)
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
	if mock.File != "" {
		if !strings.HasSuffix(mock.File, ".json") {
			return fmt.Errorf("[Route %s] mock.file must be a .json file, got '%s'", routePath, mock.File)
		}

		mockFilePath := msUtils.ResolveMockFilePath(configFilePath, mock.File)

		if _, err := os.Stat(mockFilePath); err != nil {
			return fmt.Errorf("[Route %s] mock.file not found: '%s'", routePath, mock.File)
		}
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

func validateStateful(cfg *StatefulConfig, routePath string) error {
	if cfg == nil {
		return nil
	}

	if cfg.Collection == "" {
		return fmt.Errorf("stateful route '%s' missing required field: 'collection'", routePath)
	}

	if cfg.Action == "" {
		return fmt.Errorf("stateful route '%s' missing required field: 'action'", routePath)
	}
	validActions := map[string]bool{
		"create": true, "get": true, "update": true, "delete": true, "list": true,
	}
	if !validActions[cfg.Action] {
		return fmt.Errorf("stateful route '%s' has invalid action '%s'. Valid actions: create, get, update, delete, list", routePath, cfg.Action)
	}

	return nil
}

func validateCases(cases []CaseConfig, routePath string) error {
	if len(cases) > maxCasesPerRoute {
		return fmt.Errorf("[Route %s] too many cases (%d), max allowed is %d",
			routePath, len(cases), maxCasesPerRoute)
	}

	for i, c := range cases {
		if strings.TrimSpace(c.When) == "" {
			return fmt.Errorf("[Route %s][case %d] when condition cannot be empty", routePath, i)
		}

		if err := validateConditionExpression(c.When); err != nil {
			return fmt.Errorf("[Route %s][case %d] invalid condition: %w", routePath, i, err)
		}

		if err := validateCaseResponse(&c.Then, routePath, i); err != nil {
			return err
		}
	}

	return nil
}

func validateConditionExpression(expr string) error {
	expr = strings.TrimSpace(expr)

	if len(expr) > 256 {
		return fmt.Errorf("condition too long (max 256 chars)")
	}

	// Forbidden characters control
	if strings.ContainsAny(expr, "`;$") {
		return fmt.Errorf("condition contains forbidden characters")
	}

	matches := rootRegex.FindAllString(expr, -1)

	if len(matches) == 0 {
		return fmt.Errorf(
			"condition must reference one of: body, query, headers, path, method",
		)
	}

	return nil
}

func validateCaseResponse(resp *CResponse, routePath string, index int) error {
	if resp.Status < 100 || resp.Status > 599 {
		return fmt.Errorf("[Route %s][case %d] invalid status code %d",
			routePath, index, resp.Status)
	}

	if resp.DelayMs < 0 {
		return fmt.Errorf("[Route %s][case %d] delay_ms cannot be negative",
			routePath, index)
	}

	return nil
}
