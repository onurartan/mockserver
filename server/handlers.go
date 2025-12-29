package server

import (
	"fmt"

	"bytes"
	"context"
	"io"
	"os"
	"time"

	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
	mslogger "mockserver/logger"
	msServerHandlers "mockserver/server/handlers"
	server_utils "mockserver/server/utils"
	msUtils "mockserver/utils"
)

type BaseHandlerFunc func(c *fiber.Ctx, ctx server_utils.EContext) error

// computeDelay determines the response delay based on a precedence hierarchy:
// Route Config > Global Config > Server Default.
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

var pathRegex = regexp.MustCompile(`{[a-zA-Z0-9_]+}`)

// compilePathRegex transforms OpenAPI-style path parameters (e.g., "/users/{id}")
// into Go Regex named capturing groups (e.g., "/users/(?P<id>[^/]+)") for dynamic matching.
func compilePathRegex(path string) (*regexp.Regexp, error) {
	pathRegexStr := pathRegex.ReplaceAllStringFunc(path, func(s string) string {
		name := strings.Trim(s, "{}")
		return fmt.Sprintf("(?P<%s>[^/]+)", name)
	})
	return regexp.Compile(pathRegexStr)
}

// [IMP_FUNC]
// newMockHandler initializes a MockHandler.
// It resolves configuration precedence (Status, Headers) and pre-loads mock data (Body vs File).
func newMockHandler(cfg *msconfig.MockConfig, routeCfg msconfig.RouteConfig, srvCfg msconfig.ServerConfig, configFilePath string, stateStore *server_utils.StateStore) (*MockHandler, error) {
	if routeCfg.Method != "" {
		if err := msUtils.ValidateRouteMethod(routeCfg.Method); err != nil {
			mslogger.LogError(err.Error(), 0, 0, 5)
			return nil, err
		}
	}

	// Resolve HTTP Status Code
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

	var (
		mockBodyData interface{}
		mockFileData []byte
		mockFilePath string
	)

	// Determine Data Source: Inline 'Body' takes precedence over 'File'
	if cfg.Body != nil {
		mockBodyData = cfg.Body
	} else if cfg.File != "" {
		mockFilePath = msUtils.ResolveMockFilePath(configFilePath, cfg.File)
		data, err := os.ReadFile(mockFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read mock file: %w", err)
		}
		mockFileData = data
	} else {
		return nil, fmt.Errorf("mock must define either 'body' or 'file'")
	}

	return &MockHandler{
		routeName:    routeCfg.Name,
		filePath:     mockFilePath,
		status:       status,
		headers:      headers,
		delayMs:      delay,
		mockBodyData: mockBodyData,
		mockFileData: mockFileData,
		stateStore:   stateStore,
		routecfg:     routeCfg,
	}, nil
}

// Handler executes the mock logic.
// It performs schema validation, artificial delays, and template processing (variable injection)
// before returning the final JSON response.
func (m *MockHandler) handler(c *fiber.Ctx, ctx server_utils.EContext) error {

	applyDelay(m.delayMs)

	for k, v := range m.headers {
		c.Set(k, v)
	}

	// Aggregate all parameters (Path + Query) for template substitution
	params := make(map[string]string)
	for k, v := range c.AllParams() {
		params[k] = v
	}
	for k, v := range c.Queries() {
		params[k] = v
	}

	// Parse body for Schema Validation if available
	var body map[string]interface{}
	if shouldParseBody(c) {
		if err := c.BodyParser(&body); err != nil {
			// return c.Status(400).JSON(fiber.Map{
			// 	"error": "invalid body",
			// })
			return responseError(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error(), false)
		}

		if m.routecfg.BodySchema != nil {
			// Enforce strict JSON Schema validation
			if err := server_utils.ValidateJSONSchema(m.routecfg.BodySchema, body, "request.body"); err != nil {
				return responseError(c, fiber.StatusBadRequest, "SCHEMA_VALIDATION_FAILED", err.Error(), false)
			}
		}
	} else {
		body = make(map[string]interface{})
	}

	var responseBody interface{}

	if m.mockBodyData != nil {
		// Scenario A: Process Inline Mock (Dynamic Templates supported)
		processed, err := server_utils.ProcessTemplateJSON(m.mockBodyData, ctx)
		if err != nil {
			return responseError(c, 500, "TEMPLATE_ERROR", err.Error(), false)
		}
		responseBody = processed

	} else {
		// Scenario B: Process Legacy File-based Mock (Filtering supported)
		filtered, err := parseAndFilterMockData(m.mockFileData, ctx, params)
		if err != nil {
			return responseError(c, 500, "MOCK_PARSE_ERROR", err.Error(), false)
		}
		responseBody = filtered
	}

	c.Status(m.status)
	return c.JSON(responseBody)
}

// [IMP_FUNC]
// newFetchHandler prepares a proxy handler.
// It parses the target URL and compiles path matching regexes to ensure safe proxying.
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

// Handler acts as a Reverse Proxy.
// It constructs a new downstream request, forwarding allowed headers and body,
// while enforcing timeouts and handling artificial delays.
func (p *FetchHandler) handler(c *fiber.Ctx, ctx server_utils.EContext) error {

	start := time.Now()

	timeout := 10 * time.Second
	if p.timeoutMs > 0 {
		timeout = time.Duration(p.timeoutMs) * time.Millisecond
	}

	// Create a context with timeout to prevent hanging connections
	timeCtx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()

	if p.delayMs > 0 {
		select {
		case <-time.After(time.Duration(p.delayMs) * time.Millisecond):
		// Delay completed successfully
		case <-timeCtx.Done():
			return responseError(c, fiber.StatusGatewayTimeout, "FETCH_TIMEOUT_ERROR",
				fmt.Sprintf("Request exceeded timeout of %d ms during delay", p.timeoutMs), false)
		}
	}
	method := p.method
	if method == "" {
		method = c.Method()
	}

	// Build Target URL
	pathParams := c.AllParams()
	clientQueryParams := map[string]string{}
	for k, v := range c.Queries() {
		clientQueryParams[k] = v
	}

	targetURL := buildTargetURL(p.targetURL, pathParams, clientQueryParams, p.queryParams, p.fetchQueryParams)
	mslogger.LogInfo(fmt.Sprintf("Proxying request: %s %s", method, targetURL), 0, 0, 5)

	// Prepare Request Body
	var body io.Reader
	if method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch {
		body = bytes.NewReader(c.Body())
	}

	req, err := http.NewRequestWithContext(timeCtx, method, targetURL, body)
	if err != nil {
		mslogger.LogError(fmt.Sprintf("Failed to create request: %v", err), 0, 0, 5)
		return responseError(c, fiber.StatusInternalServerError, "FETCH_BUILD_REQUEST_ERROR", err.Error(), false)
	}

	// Header Forwarding Strategy:
	// 1. Apply headers defined in FetchConfig
	// 2. Forward client headers (unless overridden by config)
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}
	c.Request().Header.VisitAll(func(key, val []byte) {
		k := string(key)
		if _, ok := p.headers[k]; !ok {
			req.Header.Set(k, string(val))
		}
	})

	// Execute Request
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

	// Metrics / Meta Data
	c.Locals(msServerHandlers.CtxUpstreamURL, targetURL)
	c.Locals(msServerHandlers.CtxUpstreamStatus, resp.StatusCode)
	c.Locals(msServerHandlers.CtxUpstreamTimeMs, time.Since(start).Milliseconds())

	// Handle 304 Not Modified transparently
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

	// Pass upstream errors to client
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return responseError(c, resp.StatusCode, "FETCH_UPSTREAM_CLIENT_ERROR", "An unknown error occurred while sending the request to the specified URL.", false)
	}

	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Set(k, v)
		}
	}

	return c.Send(bodyBytes)
}

// handleStateError maps internal storage errors to standardized HTTP API responses.
// It provides helpful hints for 404 (Not Found) and 409 (Conflict) scenarios.
func handleStateError(c *fiber.Ctx, err error, route msconfig.RouteConfig, ctx server_utils.EContext) error {
	if err == server_utils.StateErrNotFound {
		return c.Status(404).JSON(fiber.Map{
			"error": fiber.Map{
				"code":       "STATE_NOT_FOUND",
				"message":    "Item not found in collection",
				"collection": route.Stateful.Collection,
				"id":         ctx.Path[route.Stateful.IDField],
				"hint": fmt.Sprintf(
					"Ensure the item exists or create it first via POST %s",
					strings.Split(route.Path, "/{")[0],
				),
			},
		})
	}

	if err == server_utils.StateErrConflict {
		return c.Status(409).JSON(fiber.Map{
			"error": fiber.Map{
				"code":       "STATE_CONFLICT",
				"message":    "Item already exists",
				"collection": route.Stateful.Collection,
				"id":         ctx.Body[route.Stateful.IDField],
				"hint": fmt.Sprintf(
					"Use PUT %s/{id} to update the existing item",
					strings.Split(route.Path, "/{")[0],
				),
			},
		})
	}

	return responseError(c, 500, "STATE_ERROR", err.Error(), false)
}

// [IMP_FUNC]
// createRouteHandler constructs an optimized HTTP request processing pipeline for a specific route configuration.
// It performs startup-time initialization of Mock and Fetch strategies to minimize runtime overhead
// and ensure high-performance request handling.
//
// Execution Pipeline Hierarchy:
//  1. Context Initialization: Packages request metadata (headers, query, params) into the EContext structure.
//  2. Stateful Logic: If 'stateful' is enabled, executes CRUD operations on the In-memory State Engine
//     before any response logic is triggered.
//  3. Conditional Cases: Evaluates 'When/Then' priority scenarios. The first matching case terminates the
//     pipeline and returns the associated response.
//  4. Base Handler (Fallback): If no cases match, executes the pre-initialized Mock or Fetch handler.
//  5. Default Fallback: If no handler matched and a 'Default' response is defined, it serves as the final
//     result (Fetch routes are excluded from this fallback).
//
// Parameters:
//   - route: The RouteConfig object containing the specific route definition.
//   - srvCfg: Global server configuration for default delays, auth, and headers.
//   - configFilePath: Base directory path for resolving file-based mocks.
//   - stateStore: The thread-safe in-memory store for managing stateful collections.
//
// Returns:
//   - fiber.Handler: A compiled Go-Fiber handler ready for router registration.
//   - error: Returns an error if regex compilation or handler initialization fails during startup.
func createRouteHandler(route msconfig.RouteConfig, srvCfg msconfig.ServerConfig, configFilePath string, stateStore *server_utils.StateStore) (fiber.Handler, error) {

	var baseHandler BaseHandlerFunc
	var err error

	// Initialize the appropriate Base Handler (Mock or Fetch)
	if route.Mock != nil {
		var mh *MockHandler
		mh, err = newMockHandler(route.Mock, route, srvCfg, configFilePath, stateStore)
		if err != nil {
			return nil, err
		}
		baseHandler = withRouteMetaContext(
			msServerHandlers.RouteTypeMock,
			mh.routeName,
			mh.handler,
		)
	} else if route.Fetch != nil {
		var fh *FetchHandler
		fh, err = newFetchHandler(route.Fetch, route, srvCfg)
		if err != nil {
			return nil, err
		}
		baseHandler = withRouteMetaContext(
			msServerHandlers.RouteTypeFetch,
			fh.routeName,
			fh.handler,
		)
	}

	return func(c *fiber.Ctx) error {
		// Build EContext
		ctx := server_utils.EContext{
			Headers: buildHeaders(c),
			Query:   buildQuery(c),
			Path:    c.AllParams(),
			Body:    map[string]interface{}{},
		}
		if len(c.Body()) > 0 {
			json.Unmarshal(c.Body(), &ctx.Body)
		}

		// Execute Stateful Logic (if configured)
		// This handles CRUD operations on the state store before any response logic.
		if route.Stateful != nil {
			if err := server_utils.ApplyStateful(stateStore, route.Stateful, &ctx); err != nil {
				return handleStateError(c, err, route, ctx)
			}
		}

		// Evaluate Conditional Cases (Priority Logic)
		// If a "Case" matches, it returns immediately, bypassing the Base Handler.
		if len(route.Cases) > 0 {
			for _, cs := range route.Cases {
				match, err := server_utils.EvaluateCondition(cs.When, ctx)
				if err != nil {
					return responseError(c, 500, "CASE_EVAL_ERROR", err.Error(), false)
				}
				if match {
					applyDelay(cs.Then.DelayMs)
					for k, v := range cs.Then.Headers {
						c.Set(k, v)
					}
					processed, err := server_utils.ProcessTemplateJSON(cs.Then.Body, ctx)
					if err != nil {
						return responseError(c, 500, "TEMPLATE_PROCESS_ERROR", err.Error(), false)
					}
					c.Status(cs.Then.Status)
					return c.JSON(processed)
				}
			}
		}

		// Execute Base Handler (Fallback)
		if baseHandler != nil {
			return baseHandler(c, ctx)
		}

		//  Default Handler (Fallback)
		if route.Default != nil && route.Fetch == nil {
			applyDelay(route.Default.DelayMs)

			for k, v := range route.Default.Headers {
				c.Set(k, v)
			}

			processed, err := server_utils.ProcessTemplateJSON(route.Default.Body, ctx)
			if err != nil {
				return responseError(c, 500, "DEFAULT_TEMPLATE_ERROR", err.Error(), false)
			}

			c.Status(route.Default.Status)
			return c.JSON(processed)
		}

		return responseError(c, fiber.StatusNotFound, "HANDLER_NOT_MATCHED", "No handler matched", false)
	}, nil
}
