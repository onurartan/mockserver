package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

import (
	msconfig "mockserver/config"
	mslogger "mockserver/logger"
)

const (
	// Application version
	Version = "0.0.1"

	// Debounce delay for config reload
	debounceDelay = 500 * time.Millisecond
)

var configFile string

func main() {
	mslogger.StartupMessage(Version)
	mslogger.LoggerConfig.ShowTimestamp = false

	var rootCmd = &cobra.Command{
		Use:   "mockserver",
		Short: "MockServer CLI",
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the mock server",
		Run: func(cmd *cobra.Command, args []string) {
			if configFile == "" {
				fmt.Println("Config file is required. Example: mockserver start --config mockserver.json")
				os.Exit(1)
			}

			startApp(configFile)
		},
	}

	startCmd.Flags().StringVarP(&configFile, "config", "c", "mockserver.json", "Path to config file")
	rootCmd.AddCommand(startCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startApp(configFile string) {

	// Config dosyasını absolute path’e çevir
	absConfigPath, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Printf("[ERROR] Failed to resolve config path: %v\n", err)
		os.Exit(1)
	}

	// configDir := filepath.Dir(absConfigPath)

	app, cfg := mustLoadAndStart(absConfigPath)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	go listenApp(app, addr)
	mslogger.LogServerStart(addr)

	watchConfigFile(configFile, &app, &cfg)
}

func parseFlags() string {
	configFile := flag.String("config", "mockserver.json", "Path to config file (required)")
	flag.Parse()
	if *configFile == "" {
		fatalExit("Config file parameter is required. Example: mockserver -config=mockserver.json")
	}
	return *configFile
}

// watchConfigFile sets up fsnotify watcher and handles reload
func watchConfigFile(configFile string, appPtr **fiber.App, cfgPtr **msconfig.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatalExit(fmt.Sprintf("Failed to start config watcher: %v", err))
	}
	defer watcher.Close()

	if err := watcher.Add(configFile); err != nil {
		fatalExit(fmt.Sprintf("Failed to watch config file: %v", err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var reloadTimer *time.Timer
	var mu sync.Mutex

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				mu.Lock()
				if reloadTimer != nil {
					reloadTimer.Stop()
				}
				reloadTimer = time.AfterFunc(debounceDelay, func() {
					reloadServer(appPtr, cfgPtr, configFile)
				})
				mu.Unlock()
			}

		case err := <-watcher.Errors:
			mslogger.LogError(fmt.Sprintf("Config watcher error: %v", err))

		case sig := <-sigChan:
			handleSignal(sig, *appPtr)
			return
		}
	}
}

func handleSignal(sig os.Signal, app *fiber.App) {
	mslogger.LogWarn(fmt.Sprintf("Signal received (%s), shutting down gracefully...", sig))
	_ = app.Shutdown()
	mslogger.LogInfo("MockServer stopped. Goodbye! 👋")
}
