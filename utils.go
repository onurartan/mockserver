package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"

	msconfig "mockserver/config"
	mslogger "mockserver/logger"
	msServer "mockserver/server"
)




// mustLoadAndStart loads config and starts server
func mustLoadAndStart(configPath string) (*fiber.App, *msconfig.Config) {
	cfg, err := msconfig.LoadConfig(configPath)
	if err != nil {
		fatalExit(fmt.Sprintf("Failed to load config: %v", err))
	}
	return msServer.StartServer(cfg, configPath), cfg
}

// listenApp starts the Fiber server
func listenApp(app *fiber.App, addr string) {
	if err := app.Listen(addr); err != nil {
		mslogger.LogError(fmt.Sprintf("Server stopped unexpectedly: %v", err))
	}
}

// reloadServer reloads config and restarts server
func reloadServer(appPtr **fiber.App, cfgPtr **msconfig.Config, configFile string) {
	mslogger.LogWarn("Config file changed. Reloading server...")

	_ = (*appPtr).Shutdown()
	time.Sleep(200 * time.Millisecond) // short wait to release port

	newCfg, err := msconfig.LoadConfig(configFile)
	if err != nil {
		mslogger.LogError(fmt.Sprintf("Failed to reload config: %v", err))
		return
	}

	newApp := msServer.StartServer(newCfg, configFile)
	newAddr := fmt.Sprintf(":%d", newCfg.Server.Port)
	go listenApp(newApp, newAddr)
	mslogger.LogSuccess(fmt.Sprintf("Server reloaded successfully and listening on %s", mslogger.GetServerHost(newAddr)), 1)

	*appPtr = newApp
	*cfgPtr = newCfg
}

// fatalExit logs error and exits
func fatalExit(msg string) {
	mslogger.LogError(msg)
	os.Exit(1)
}
