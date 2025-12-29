package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"

	msconfig "mockserver/config"
	mslogger "mockserver/logger"
	msServer "mockserver/server"
)


func mustLoadAndStart(configPath string) *Runtime {
	cfg, err := msconfig.LoadConfig(configPath)
	if err != nil {
		fatalExit(fmt.Sprintf("Failed to load config: %v", err))
	}

	app := msServer.StartServer(cfg, configPath, embedDir)

	return &Runtime{
		App: app,
		Cfg: cfg,
	}
}

// listenApp starts the Fiber server
func listenApp(app *fiber.App, addr string) {
	if err := app.Listen(addr); err != nil {
		mslogger.LogError(fmt.Sprintf("Server stopped unexpectedly: %v", err))
	}
}


func reloadServer(configFile string, rt *Runtime) {
	rt.Mu.Lock()
	defer rt.Mu.Unlock()

	mslogger.LogWarn("Config file changed. Reloading server...")

	cfg, err := msconfig.LoadConfig(configFile)
	if err != nil {
		mslogger.LogError("Reload failed: " + err.Error())
		return
	}

	// close old server
	if rt.App != nil {
		_ = rt.App.Shutdown()
	}

	newApp := msServer.StartServer(cfg, configFile, embedDir)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	go listenApp(newApp, addr)

	rt.App = newApp
	rt.Cfg = cfg

	mslogger.LogSuccess(
		fmt.Sprintf("Server reloaded and listening on %s", mslogger.GetServerHost(addr, "")),
		1,
	)
}


// fatalExit logs error and exits
func fatalExit(msg string) {
	mslogger.LogError(msg)
	os.Exit(1)
}
