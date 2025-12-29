package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
	msServerHandlers "mockserver/server/handlers"
	server_utils "mockserver/server/utils"
)

// validateDelay ensures the artificial delay does not exceed the safety limit (10 seconds).
// This prevents configuration errors from causing long-hanging connections.
func validateDelay(delay int) (int, error) {
	if delay > 10000 {
		return 0, fmt.Errorf("delay cannot exceed 10000 ms (10 seconds), got %d", delay)
	}
	return delay, nil
}

// mergeHeaders combines HTTP headers from multiple sources with a specific precedence order:
// Default Config < Route Config < Custom Overrides.
// Keys in later maps overwrite keys in earlier maps.
func mergeHeaders(defaults, routeHeaders, customHeaders map[string]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range defaults {
		headers[k] = v
	}
	for k, v := range routeHeaders {
		headers[k] = v
	}
	for k, v := range customHeaders {
		headers[k] = v
	}
	return headers
}

func applyDelay(ms int) {
	if ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

// buildHeaders extracts and normalizes all request headers into a simple map.
// Header keys are converted to lowercase for consistent case-insensitive lookups.
func buildHeaders(c *fiber.Ctx) map[string]string {
	h := make(map[string]string)
	for k, v := range c.GetReqHeaders() {
		if len(v) > 0 {
			h[strings.ToLower(k)] = v[0]
		}
	}
	return h
}

// buildQuery extracts all query parameters into a map, normalizing keys to lowercase.
func buildQuery(c *fiber.Ctx) map[string]string {
	q := make(map[string]string)
	for k, v := range c.Queries() {
		q[strings.ToLower(k)] = v
	}
	return q
}

// shouldParseBody determines if the HTTP method typically supports a request body.
func shouldParseBody(c *fiber.Ctx) bool {
	switch c.Method() {
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch:
		return len(c.Body()) > 0
	default:
		return false
	}
}

// parseAndFilterMockData processes raw JSON templates and applies filtering logic.
// 1. Unmarshals raw bytes into a generic interface.
// 2. Executes template substitution (e.g., {{fake.Name}}).
// 3. Normalizes single objects into a slice of objects.
// 4. Applies query parameter filtering to the result set.
func parseAndFilterMockData(data []byte, ctx server_utils.EContext, params map[string]string) ([]map[string]interface{}, error) {

	var rawData interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	processed, err := server_utils.ProcessTemplateJSON(rawData, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to process template JSON: %w", err)
	}
	var arr []interface{}

	// Normalize data structure: Ensure we always work with a slice
	switch v := processed.(type) {
	case []interface{}:
		arr = v
	case map[string]interface{}:
		// Wrap single object in array
		arr = []interface{}{v}
	default:
		return nil, fmt.Errorf("mock data must be an object or array of objects")
	}

	// Type assertion for elements
	result := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("mock array items must be objects")
		}
		result = append(result, m)
	}

	filtered, err := server_utils.FilteredMockData(result, params)
	if err != nil {
		return nil, fmt.Errorf("failed to filter mock data: %w", err)
	}
	return filtered, nil
}

// buildTargetURL constructs the final upstream URL for proxy requests.
// It handles path parameter substitution (e.g., {id} -> 123) and merges
// client query parameters with configured overrides.
func buildTargetURL(base *url.URL, pathParams, clientQuery map[string]string, acceptedQueryParams map[string]struct{}, fetchQueryParams map[string]string) string {
	target := *base

	// Path Parameter Substitution
	path := target.Path
	for k, v := range pathParams {
		path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", k), v)
	}
	target.Path = path

	// Forward allowed client params
	q := target.Query()
	for k, v := range clientQuery {
		if _, ok := acceptedQueryParams[k]; ok {
			q.Set(k, v)
		}
	}

	for k, v := range fetchQueryParams {
		q.Set(k, v)
	}

	target.RawQuery = q.Encode()
	return target.String()
}

// responseError writes a standardized JSON error response to the client.
// It optionally returns the ApiError struct for internal error handling flows.
func responseError(c *fiber.Ctx, status int, errCode, message string, returnObject bool) error {
	apiErr := &ApiError{
		Success:   false,
		Status:    status,
		Err:       http.StatusText(status),
		ErrorCode: errCode,
		Message:   message,
		Timestamp: time.Now().UTC().UnixNano() / 1e6,
	}

	err := c.Status(status).JSON(apiErr)

	if returnObject {
		return apiErr
	}

	return err
}

// getRoutesStat calculates summary statistics for the registered routes.
// Returns (Total Routes, Mock Routes, Fetch Routes).
func getRoutesStat(cfg *msconfig.Config) (int, int, int) {
	routeCount := 0
	mockCount := 0
	fetchCount := 0

	for _, route := range cfg.Routes {
		routeCount++

		if route.Mock != nil {
			mockCount++
		}
		if route.Fetch != nil {
			fetchCount++
		}
	}

	return routeCount, mockCount, fetchCount
}

// withRouteMeta decorates a standard Fiber Handler with route metadata.
// Used for injecting context into logs and middleware (e.g., route name, type).
func withRouteMeta(
	routeType string,
	routeName string,
	handler fiber.Handler,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(msServerHandlers.CtxRouteType, routeType)
		c.Locals(msServerHandlers.CtxRouteName, routeName)
		c.Locals(msServerHandlers.CtxRoutePath, strings.Split(c.OriginalURL(), "?")[0]+"")
		return handler(c)
	}
}

// withRouteMetaContext decorates a BaseHandlerFunc (which includes EContext).
// This is the Context-Aware version of withRouteMeta for Mock/Fetch handlers.
func withRouteMetaContext(
	routeType string,
	routeName string,
	handler BaseHandlerFunc,
) BaseHandlerFunc {
	return func(c *fiber.Ctx, ctx server_utils.EContext) error {
		c.Locals(msServerHandlers.CtxRouteType, routeType)
		c.Locals(msServerHandlers.CtxRouteName, routeName)
		c.Locals(msServerHandlers.CtxRoutePath, strings.Split(c.OriginalURL(), "?")[0]+"")
		return handler(c, ctx)
	}
}
