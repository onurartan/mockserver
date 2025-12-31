package server

import (
	"fmt"
	"io/fs"
	"net/http"

	// "os"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

import (
	msconfig "mockserver/config"
	mslogger "mockserver/logger"
	appinfo "mockserver/pkg/appinfo"
	msServerHandlers "mockserver/server/handlers"
	server_utils "mockserver/server/utils"
	msUtils "mockserver/utils"
)

var idRegex = regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)

// GlobalStateStore holds the in-memory state for stateful routes.
// It is initialized once at startup.
var globalStateStore = server_utils.NewStateStore()

func (e *ApiError) Error() string {
	return e.Message
}

// StartServer initializes and configures the Fiber application.
//
// It orchestrates the following bootstrap process:
// 1. Configures the Fiber app engine (Error handling, 405 behaviors).
// 2. Registers global middleware (CORS, Recovery, Logging).
// 3. Mounts internal endpoints (Console, Swagger, Debug).
// 4. Compiles and registers user-defined routes.
//
// Returns the configured *fiber.App instance ready for listening.
func StartServer(cfg *msconfig.Config, configFilePath string, embedFS fs.FS, faviconFS fs.FS) *fiber.App {

	// Initialize background log aggregation
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

	// Middleware
	setupMiddleware(app, cfg, faviconFS)

	// ConsoleUI
	SetupConsoleRoutes(app, cfg, embedFS)

	// OpenAPI / Swagger UI
	app.Get("/openapi.json", func(c *fiber.Ctx) error {
		openapi := generateOpenAPISpec(cfg)
		return c.JSON(openapi)
	})
	app.Get(cfg.Server.SwaggerUIPath, swaggerUIHandler)

	// Debug Routes
	if cfg.Server.Debug != nil && cfg.Server.Debug.Enabled {
		setupDebugRoutes(app, cfg)
	}
	// Register User Routes
	registerUserRoutes(app, cfg, configFilePath)

	// Fallback Handler (404)
	app.Use(RegisterFallback())

	return app
}

// setupMiddleware attaches global middleware to the Fiber app.
func setupMiddleware(app *fiber.App, cfg *msconfig.Config, faviconFS fs.FS) {
	// Favicon
	// if _, err := os.Stat("./favicon.ico"); err == nil {
		app.Use(favicon.New(favicon.Config{
			FileSystem: http.FS(faviconFS),
			File:       "favicon.ico",
			URL:        "/favicon.ico",
		}))
	// }

	// Panic Recovery
	app.Use(recover.New())

	// Request Logging (Custom)
	app.Use(msServerHandlers.RequestLoggerMiddleware(cfg.Server.Debug.Path, cfg))

	// CORS
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

	// Console/Debug Exclusion Logger
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		// Skip logging for internal dashboard paths to keep logs clean
		if msServerHandlers.IgnoredPaths[c.Path()] ||
			strings.HasPrefix(c.Path(), cfg.Server.Console.Path) ||
			strings.HasPrefix(c.Path(), cfg.Server.Debug.Path) {
			return nil
		}
		mslogger.LogRoute(c.Method(), c.Path(), c.IP(), c.Response().StatusCode(), duration, "    ")
		return err
	})
}

// registerUserRoutes iterates over the configuration and registers endpoints.
// It normalizes API prefixes and path parameters (converting {id} to :id).
func registerUserRoutes(app *fiber.App, cfg *msconfig.Config, configFilePath string) {
	prefix := normalizePrefix(cfg.Server.APIPrefix)

	maxLogRoutes := 10
	routeLogCount := 0

	for _, route := range cfg.Routes {
		handler, err := createRouteHandler(route, cfg.Server, configFilePath, globalStateStore)
		if err != nil {
			msUtils.StopWithError(fmt.Sprintf("Failed to create route: %s", route.Name), err)
			continue
		}

		// Convert OpenAPI style path "{id}" to Fiber style ":id"
		fiberPath := idRegex.ReplaceAllString(route.Path, `:$1`)
		routePath := prefix + fiberPath
		method := strings.ToUpper(route.Method)

		// Register the specific method
		registerRoute(app, method, routePath, authMiddleware(cfg.Server.Auth, route.Auth), handler)

		// Logging
		routeLogCount++
		if routeLogCount <= maxLogRoutes {
			mslogger.LogRoute(method, routePath, "", 0, 0, "[ROUTE_REGISTERED]")
		}
	}

	if len(cfg.Routes) > maxLogRoutes {
		mslogger.LogInfo(fmt.Sprintf("+%d more routes registered...", len(cfg.Routes)-maxLogRoutes))
	}
}

// registerRoute is a helper to dynamically register handlers based on string method names.
func registerRoute(app *fiber.App, method, path string, mw, handler fiber.Handler) {
	switch strings.ToUpper(method) {
	case fiber.MethodGet:
		app.Get(path, mw, handler)
	case fiber.MethodPost:
		app.Post(path, mw, handler)
	case fiber.MethodPut:
		app.Put(path, mw, handler)
	case fiber.MethodPatch:
		app.Patch(path, mw, handler)
	case fiber.MethodDelete:
		app.Delete(path, mw, handler)
	}
}

// Debug route'ları ayırmak için (Opsiyonel temizlik)
func setupDebugRoutes(app *fiber.App, cfg *msconfig.Config) {
	debugRequestPath := cfg.Server.Debug.Path + "/requests"
	debugHealthPath := cfg.Server.Debug.Path + "/health"

	app.Get(debugRequestPath, withRouteMeta(msServerHandlers.RouteTypeInternal, "debug_requests", msServerHandlers.DebugRequestsHandler))

	routeCount, mockCount, fetchCount := getRoutesStat(cfg)
	app.Get(debugHealthPath, withRouteMeta(msServerHandlers.RouteTypeInternal, "debug_health",
		msServerHandlers.HealthHandler(routeCount, mockCount, fetchCount, appinfo.Version)))
}

func normalizePrefix(prefix string) string {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if strings.HasSuffix(prefix, "/") {
		prefix = strings.TrimSuffix(prefix, "/")
	}
	return prefix
}
