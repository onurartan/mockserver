package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	msconfig "mockserver/config"
)

// If the Mockserver Console UI is active, it configures the route settings.
func SetupConsoleRoutes(app *fiber.App, cfg *msconfig.Config, embedFS fs.FS) {

	initJWTSecret(cfg)

	if !cfg.Server.Console.Enabled {
		return
	}

	consoleCfg := cfg.Server.Console
	cPath := strings.TrimRight(consoleCfg.Path, "/")

	appFS, err := fs.Sub(embedFS, "www")
	if err != nil {
		panic("Embed FS Error: " + err.Error())
	}

	// Static Assets
	publicFS, _ := fs.Sub(appFS, "public")
	app.Use("/public", filesystem.New(filesystem.Config{
		Root:       http.FS(publicFS),
		// PathPrefix: "public",
		Browse:     false,
	}))

	// Console UI Login Settings
	app.Get(cPath+"/login", func(c *fiber.Ctx) error {

		token := c.Cookies(JWTCookieName)
		if token != "" {
			if _, err := validateToken(token); err == nil {
				return c.Redirect(cPath)
			}
		}

		content, _ := fs.ReadFile(appFS, "login.html")
		c.Set("Content-Type", "text/html")

		return c.Send(content)
	})

	app.Post(cPath+"/login", ConsoleLoginHandler(cfg))

	consoleGroup := app.Group(cPath, ConsoleAuthMiddleware(cfg))

	// JS + CSS Route Settings
	consoleAssets := consoleGroup.Group("/", ConsoleAssetGuard(consoleCfg))
	jsFS, _ := fs.Sub(appFS, "js")
	consoleAssets.Group("/js").Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(jsFS),
		Browse: false,
	}))
	cssFS, _ := fs.Sub(appFS, "css")
	consoleAssets.Group("/css").Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(cssFS),
		Browse: false,
	}))

	// Console UI main Settings
	consoleGroup.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		content, err := fs.ReadFile(appFS, "index.html")
		if err != nil {
			return c.Status(500).SendString("System Error: Index missing")
		}
		return c.Send(content)
	})

	// Other Endpoints
	consoleGroup.Get("/me", ConsoleMeHandler)
	consoleGroup.Get("/mockserver.json", SafeConfigHandler(cfg))
	consoleGroup.Get("/logout", ConsoleLogoutHandler(cfg))
}
