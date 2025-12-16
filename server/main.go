package server

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"net/http"
)

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

import (
	msconfig "mockserver/config"
	appinfo "mockserver/internal/appinfo"
	mslogger "mockserver/logger"
	msServerHandlers "mockserver/server/handlers"
	msUtils "mockserver/utils"
)

func (e *ApiError) Error() string {
	return e.Message
}

// [IMP_FUNC]
// StartServer initializes a Fiber HTTP server with middleware, error handling, logging, CORS,
// authentication, and routes based on the provided configuration.
//
// The returned *fiber.App is ready to listen. It handles:
//   - Structured API errors with consistent JSON responses
//   - Panic recovery and request logging
//   - Optional CORS and authentication per route
//   - Swagger/OpenAPI endpoints
//   - Automatic 404 responses for unmatched routes
//
// Routes are registered using `createRouteHandler` and support GET, POST, PUT, PATCH, DELETE.
// Path parameters in {param} style are converted to Fiber's :param format.
func StartServer(cfg *msconfig.Config, configFilePath string) *fiber.App {

	// Start the request log aggregator goroutine
	msServerHandlers.StartLogAggregator()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,

		// Custom Fiber Error Handler
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"
			errorCode := "INTERNAL_SERVER_ERROR"

			if e, ok := err.(*ApiError); ok {
				code = e.Status
				message = e.Message
				errorCode = e.ErrorCode
			} else if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
				errorCode = strings.ToUpper(strings.ReplaceAll(message, " ", "_"))
			} else {
				message = fmt.Sprintf("%v", err)
				errorCode = strings.ToUpper(strings.ReplaceAll(message, " ", "_"))
			}

			apiErr := &ApiError{
				Success:   false,
				Status:    code,
				Err:       http.StatusText(code),
				ErrorCode: errorCode,
				Message:   message,
				Timestamp: time.Now().UTC().UnixNano() / 1e6,
			}
			return c.Status(code).JSON(apiErr)
		},
	})

	// >_ Middleware
	app.Use(favicon.New(favicon.Config{
		File: "./favicon.ico",
		URL:  "/favicon.ico",
	}))
	app.Use(recover.New())

	app.Use(msServerHandlers.RequestLoggerMiddleware(cfg.Server.Debug.Path))

	if cfg.Server.CORS.Enabled {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     strings.Join(cfg.Server.CORS.AllowOrigins, ","),
			AllowMethods:     strings.Join(cfg.Server.CORS.AllowMethods, ","),
			AllowHeaders:     strings.Join(cfg.Server.CORS.AllowHeaders, ","),
			AllowCredentials: cfg.Server.CORS.AllowCredentials,
		}))
	} else {
		app.Use(cors.New())
	}

	if cfg.Server.Debug != nil && cfg.Server.Debug.Enabled {

		debugRequestPath := cfg.Server.Debug.Path + "/requests"
		debugHealthPath := cfg.Server.Debug.Path + "/health"
		mslogger.LogRoute("GET", debugRequestPath, "", 0, 0, "[DEBUG_ROUTE_REGISTERED]")
		mslogger.LogRoute("GET", debugHealthPath, "", 0, 0, "[DEBUG_ROUTE_REGISTERED]")

		app.Get(
			debugRequestPath,
			withRouteMeta(
				msServerHandlers.RouteTypeInternal,
				"debug_requests",
				msServerHandlers.DebugRequestsHandler,
			),
		)

		routeCount, mockCount, fetchCount := getRoutesStat(cfg)
		app.Get(
			debugHealthPath,
			withRouteMeta(
				msServerHandlers.RouteTypeInternal,
				"debug_health",
				msServerHandlers.HealthHandler(routeCount, mockCount, fetchCount, appinfo.Version),
			),
		)
	}

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		mslogger.LogRoute(c.Method(), c.Path(), c.IP(), c.Response().StatusCode(), duration, "    ")
		return err
	})

	// OpenAPI / Swagger UI
	app.Get("/openapi.json", func(c *fiber.Ctx) error {
		openapi := generateOpenAPISpec(cfg)
		return c.JSON(openapi)
	})
	app.Get(cfg.Server.SwaggerUIPath, swaggerUIHandler)

	// Routes
	prefix := cfg.Server.APIPrefix
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}

	maxLogRoutes := 10             // Maximum number of routes to be displayed in the log
	totalRoutes := len(cfg.Routes) // Total number of routes defined in the configuration
	routeLogCount := 0             // Displays the number of logged routes.

	idRegex := regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)

	for _, route := range cfg.Routes {
		handler, err := createRouteHandler(route, cfg.Server, configFilePath)
		if err != nil {
			msUtils.StopWithError(fmt.Sprintf("Failed to create route: %s", route.Name), err)
			continue
		}

		fiberPath := idRegex.ReplaceAllString(route.Path, `:$1`)
		routePath := prefix + fiberPath

		switch strings.ToUpper(route.Method) {
		case fiber.MethodGet:
			app.Get(routePath, authMiddleware(cfg.Server.Auth, route.Auth), handler)
		case fiber.MethodPost:
			app.Post(routePath, authMiddleware(cfg.Server.Auth, route.Auth), handler)
		case fiber.MethodPut:
			app.Put(routePath, authMiddleware(cfg.Server.Auth, route.Auth), handler)
		case fiber.MethodPatch:
			app.Patch(routePath, authMiddleware(cfg.Server.Auth, route.Auth), handler)
		case fiber.MethodDelete:
			app.Delete(routePath, authMiddleware(cfg.Server.Auth, route.Auth), handler)
		default:
			mslogger.LogWarn(fmt.Sprintf("Unsupported HTTP method. method=%s, route=%s", route.Method, route.Name), 0, 0, 5)
			continue
		}

		routeLogCount++
		if routeLogCount <= maxLogRoutes {
			mslogger.LogRoute(strings.ToUpper(route.Method), routePath, "", 0, 0, "[ROUTE_REGISTERED]")
		}
	}

	if totalRoutes > maxLogRoutes {
		mslogger.LogInfo(fmt.Sprintf("+%d more routes registered...", totalRoutes-maxLogRoutes))
	}

	// Return an error if the requested page does not exist
	app.Use(func(c *fiber.Ctx) error {
		return responseError(
			c,
			fiber.StatusNotFound,
			"404_NOT_FOUND",
			fmt.Sprintf("The route %s %s does not exist", c.Method(), c.Path()),
			false,
		)
	})

	return app
}
