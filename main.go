package main

import (
	// "flag"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

		"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

import (
	appinfo "mockserver/pkg/appinfo"
	mslogger "mockserver/logger"
)

//go:embed www
var embedDir embed.FS

const (
	// Debounce delay for config reload
	debounceDelay = 500 * time.Millisecond
)

var configFile string

func main() {
	mslogger.StartupMessage(appinfo.Version)
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
	rootCmd.AddCommand(convertCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startApp(configFile string) {

	absConfigPath, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Printf("[ERROR] Failed to resolve config path: %v\n", err)
		os.Exit(1)
	}

	rt := mustLoadAndStart(absConfigPath)

	addr := fmt.Sprintf(":%d", rt.Cfg.Server.Port)
	go listenApp(rt.App, addr)
	mslogger.LogServerStart(addr)
	mslogger.LogSuccess(fmt.Sprintf("Interface: %s", mslogger.GetServerHost(addr, rt.Cfg.Server.Console.Path)), 0)

	watchConfigFile(configFile, rt)
}


// watchConfigFile sets up fsnotify watcher and handles reload
func watchConfigFile(configFile string, rt *Runtime) {
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
					reloadServer(configFile, rt)
				})
				mu.Unlock()
			}

		case err := <-watcher.Errors:
			mslogger.LogError(fmt.Sprintf("Config watcher error: %v", err))

		case sig := <-sigChan:
			handleSignal(sig, rt)
			return
		}
	}
}


func handleSignal(sig os.Signal, rt *Runtime) {
	rt.Mu.Lock()
	defer rt.Mu.Unlock()

	mslogger.LogWarn(
		fmt.Sprintf("Signal received (%s), shutting down gracefully...", sig),
	)

	if rt.App != nil {
		_ = rt.App.Shutdown()
	}

	mslogger.LogInfo("MockServer stopped. Goodbye! 👋")
}
