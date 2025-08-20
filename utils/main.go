package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

import (
	mslogger "mockserver/logger"
)

var AllowedMethods = map[string]struct{}{
	"GET": {}, "POST": {}, "PUT": {}, "PATCH": {}, "DELETE": {}, "OPTIONS": {},
}

// Checks if the provided HTTP method is valid.
func ValidateRouteMethod(method string) error {
	method = strings.ToUpper(method)
	if _, ok := AllowedMethods[method]; !ok {
		return fmt.Errorf("invalid HTTP method '%s' in route config", method)
	}
	return nil
}

// Used to stop the server or application in the event of a critical error.
func StopWithError(msg string, err error) {
	if err != nil {
		mslogger.LogError(fmt.Sprintf("%s: %v", msg, err))
	} else {
		mslogger.LogError(msg)
	}
	mslogger.LogInfo("Shutting down MockServer due to critical error. Goodbye! 👋")
	os.Exit(1)
}

func ResolveMockFilePath(configFilePath, filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}

	configDir := filepath.Dir(configFilePath) // directory where the config file is located

	mockFilePath := filepath.Join(configDir, filePath)
	return mockFilePath
}
