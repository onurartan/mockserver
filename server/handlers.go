package server

import (
	"fmt"

	"bytes"
	"context"
	"io"
	"os"
	"time"

	"errors"
	"regexp"
	"strings"

	"net/http"
	"net/url"
)

import (
	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
	mslogger "mockserver/logger"
	msServerHandlers "mockserver/server/handlers"
	msUtils "mockserver/utils"
)

func computeDelay(routeDelay, cfgDelay, defaultDelay int) (int, error) {
	delay := defaultDelay
	if routeDelay != 0 {
		delay = routeDelay
	}
	if cfgDelay != 0 {
		delay = cfgDelay
	}
	if delay < 0 {
		return 0, fmt.Errorf("invalid delay value: %d", delay)
	}
	return delay, nil
}

// Converts path parameters like {id} into a regex for matching.
func compilePathRegex(path string) (*regexp.Regexp, error) {
	pathRegexStr := regexp.MustCompile(`{[a-zA-Z0-9_]+}`).ReplaceAllStringFunc(path, func(s string) string {
		name := strings.Trim(s, "{}")
		return fmt.Sprintf("(?P<%s>[^/]+)", name)
	})
	return regexp.Compile(pathRegexStr)
}

// [IMP_FUNC]
// newMockHandler creates a MockHandler instance by validating config and reading mock data.
func newMockHandler(cfg *msconfig.MockConfig, routeCfg msconfig.RouteConfig, srvCfg msconfig.ServerConfig, configFilePath string) (*MockHandler, error) {
	if routeCfg.Method != "" {
		if err := msUtils.ValidateRouteMethod(routeCfg.Method); err != nil {
			mslogger.LogError(err.Error(), 0, 0, 5)
			return nil, err
		}
	}

	status := 200
	if routeCfg.Status != 0 {
		status = routeCfg.Status
	}
	if cfg.Status != 0 {
		status = cfg.Status
	}

	headers := mergeHeaders(srvCfg.DefaultHeaders, routeCfg.Headers, cfg.Headers)

	delay, err := computeDelay(routeCfg.DelayMs, cfg.DelayMs, srvCfg.DefaultDelayMs)
	if err != nil {
		return nil, err
	}

	// data, err := os.ReadFile(cfg.File)
	mockFilePath := msUtils.ResolveMockFilePath(configFilePath, cfg.File)
	data, err := os.ReadFile(mockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mock file: %w", err)
	}

	return &MockHandler{
		routeName: routeCfg.Name,
		filePath:  mockFilePath,
		status:    status,
		headers:   headers,
		delayMs:   delay,
		data:      data,
	}, nil
}

// Handler processes incoming HTTP requests for the mock route.
// Applies delay, sets headers, filters JSON data based on parameters,
// and returns the resulting JSON array or an error response.
func (m *MockHandler) handler(c *fiber.Ctx) error {

	if m.delayMs > 0 {
		time.Sleep(time.Duration(m.delayMs) * time.Millisecond)
	}

	for k, v := range m.headers {
		c.Set(k, v)
	}

	params := map[string]string{}
	for k, v := range c.AllParams() {
		params[k] = v
	}
	for k, v := range c.Queries() {
		params[k] = v
	}

	filtered, err := parseAndFilterMockData(m.data, params)
	if err != nil {
		mslogger.LogError(fmt.Sprintf("MockHandler error: %v", err), 0, 0, 5)
		return responseError(c, fiber.StatusInternalServerError, "MOCK_PARSE_ERROR", err.Error(), false)
	}

	if len(filtered) == 0 {
		mslogger.LogWarn(fmt.Sprintf("No records found for parameters: %v", params), 0, 0, 5)

		return c.JSON([]interface{}{})

		// [Alternative=Throwing a bug]:
		// return responseError(c, fiber.StatusNotFound, "MOCK_NO_RECORDS", "No matching records found", false)
	}

	c.Status(m.status)
	return c.JSON(filtered)
}

// [IMP_FUNC]
// newFetchHandler creates a FetchHandler based on route configuration.
// It compiles path regex for dynamic path parameters and validates delay.
// Returns an error if URL parsing or regex compilation fails.
func newFetchHandler(cfg *msconfig.FetchConfig, routeCfg msconfig.RouteConfig, srvCfg msconfig.ServerConfig) (*FetchHandler, error) {
	if routeCfg.Method != "" {
		if err := msUtils.ValidateRouteMethod(routeCfg.Method); err != nil {
			mslogger.LogError(err.Error(), 0, 0, 5)
			return nil, err
		}
	}

	parsedURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fetch URL: %w", err)
	}

	delay, err := computeDelay(routeCfg.DelayMs, cfg.DelayMs, srvCfg.DefaultDelayMs)
	if err != nil {
		return nil, err
	}

	urlRegex, err := compilePathRegex(routeCfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to compile path param regex: %w", err)
	}

	queryParams := make(map[string]struct{})
	for k := range routeCfg.Query {
		queryParams[k] = struct{}{}
	}

	return &FetchHandler{
		routeName:        routeCfg.Name,
		targetURL:        parsedURL,
		method:           cfg.Method,
		headers:          cfg.Headers,
		fetchQueryParams: cfg.QueryParams,
		queryParams:      queryParams,
		passStatus:       cfg.PassStatus,
		delayMs:          delay,
		timeoutMs:        cfg.TimeoutMs,
		urlRegex:         urlRegex,
		basePath:         routeCfg.Path,
	}, nil
}

// Handler sends the proxied request to the target URL with proper headers and body.
// Handles optional delay, timeout, and response status management.
// Logs errors and returns appropriate HTTP status codes for failures.
func (p *FetchHandler) handler(c *fiber.Ctx) error {

	start := time.Now()

	timeout := 10 * time.Second
	if p.timeoutMs > 0 {
		timeout = time.Duration(p.timeoutMs) * time.Millisecond
	}

	// Create context with timeout for both delay and HTTP request
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()

	if p.delayMs > 0 {
		select {
		case <-time.After(time.Duration(p.delayMs) * time.Millisecond):
		// delay finished
		case <-ctx.Done():
			return responseError(c, fiber.StatusGatewayTimeout, "FETCH_TIMEOUT_ERROR",
				fmt.Sprintf("Request exceeded timeout of %d ms during delay", p.timeoutMs), false)
		}
	}
	method := p.method
	if method == "" {
		method = c.Method()
	}

	pathParams := c.AllParams()
	clientQueryParams := map[string]string{}
	for k, v := range c.Queries() {
		clientQueryParams[k] = v
	}

	targetURL := buildTargetURL(p.targetURL, pathParams, clientQueryParams, p.queryParams, p.fetchQueryParams)
	mslogger.LogInfo(fmt.Sprintf("Proxying request: %s %s", method, targetURL), 0, 0, 5)

	var body io.Reader
	if method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch {
		body = bytes.NewReader(c.Body())
	}

	// HTTP request
	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		mslogger.LogError(fmt.Sprintf("Failed to create request: %v", err), 0, 0, 5)
		return responseError(c, fiber.StatusInternalServerError, "FETCH_BUILD_REQUEST_ERROR", err.Error(), false)
	}

	// Set headers from FetchConfig, then copy remaining client headers
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}
	c.Request().Header.VisitAll(func(key, val []byte) {
		k := string(key)
		if _, ok := p.headers[k]; !ok {
			req.Header.Set(k, string(val))
		}
	})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			return responseError(c, fiber.StatusGatewayTimeout, "FETCH_TIMEOUT_ERROR",
				fmt.Sprintf("Request exceeded timeout of %d ms", p.timeoutMs), false)
		}

		mslogger.LogError(fmt.Sprintf("Request failed: %v", err), 0, 0, 5)

		return responseError(c, fiber.StatusBadGateway, "FETCH_UPSTREAM_ERROR", err.Error(), false)

	}
	defer resp.Body.Close()

	c.Locals(msServerHandlers.CtxUpstreamURL, targetURL)
	c.Locals(msServerHandlers.CtxUpstreamStatus, resp.StatusCode)
	c.Locals(msServerHandlers.CtxUpstreamTimeMs, time.Since(start).Milliseconds())

	// 304 Not Modified handling
	if resp.StatusCode == http.StatusNotModified {
		mslogger.LogInfo("Upstream returned 304 Not Modified", 0, 0, 5)
		c.Status(fiber.StatusNotModified)
		return c.JSON(fiber.Map{})
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		mslogger.LogError(fmt.Sprintf("Failed to read response body: %v", err), 0, 0, 5)
		return responseError(c, fiber.StatusInternalServerError, "FETCH_BODY_READ_ERROR", err.Error(), false)
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return responseError(c, resp.StatusCode, "FETCH_UPSTREAM_CLIENT_ERROR", string(bodyBytes), false)
	}

	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Set(k, v)
		}
	}

	return c.Send(bodyBytes)
}

// [IMP_FUNC]
// createRouteHandler constructs the Fiber handler for a route.
// It decides whether the route uses a Mock or Fetch handler, initializes it,
// and wraps it with request validation if path parameters, query params, or headers are defined.
// Returns an error if the route configuration is invalid or both mock and fetch are set.
func createRouteHandler(route msconfig.RouteConfig, srvCfg msconfig.ServerConfig, configFilePath string) (fiber.Handler, error) {
	if route.Mock != nil && route.Fetch != nil {
		return nil, fmt.Errorf("a route cannot be both 'mock' and 'fetch'")
	}

	var baseHandler fiber.Handler
	if route.Mock != nil {
		mh, err := newMockHandler(route.Mock, route, srvCfg, configFilePath)
		if err != nil {
			return nil, err
		}
		baseHandler = withRouteMeta(
			msServerHandlers.RouteTypeMock,
			mh.routeName,
			mh.handler,
		)
	} else if route.Fetch != nil {
		fh, err := newFetchHandler(route.Fetch, route, srvCfg)
		if err != nil {
			return nil, err
		}
		baseHandler = withRouteMeta(
			msServerHandlers.RouteTypeFetch,
			fh.routeName,
			fh.handler,
		)
	} else {
		return nil, fmt.Errorf("route definition contains neither 'mock' nor 'fetch'")
	}

	// Wrap with validation if defined
	if len(route.PathParams) > 0 || len(route.Query) > 0 || len(route.RequestHeaders) > 0 {
		validator := validateRequestParams(route)
		return func(c *fiber.Ctx) error {
			validate_response := validator(c)
			if validate_response != nil {
				return validate_response
			}
			return baseHandler(c)
		}, nil
	}

	return baseHandler, nil
}
