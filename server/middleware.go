package server

import "strings"

import (
	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
)

// AuthMiddleware validates requests based on AuthConfig (global or route-specific)
// Route-level config overrides global config.
// Supports API Key (header/query) and Bearer token authentication.
func authMiddleware(globalAuth, routeAuth *msconfig.AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := globalAuth
		if routeAuth != nil {
			auth = routeAuth
		}

		if auth == nil || !auth.Enabled {
			return c.Next()
		}

		// Validate required fields
		if auth.Type == "" || auth.In == "" || auth.Name == "" {
			return responseError(c, fiber.StatusInternalServerError, "AUTH_MISCONFIGURED", "Authentication misconfigured", false)
		}

		var credential string
		switch strings.ToLower(auth.In) {
		case "header":
			credential = c.Get(auth.Name)
		case "query":
			credential = c.Query(auth.Name)
		default:
			return responseError(c, fiber.StatusInternalServerError, "UNSUPPORTED_AUTH_LOCATION", "Unsupported auth location", false)
		}

		if credential == "" {

			return responseError(c, fiber.StatusUnauthorized, "MISSING_CREDENTIAL", "Missing authentication credential", false)
		}

		// Validate credential
		switch strings.ToLower(auth.Type) {
		case "apikey":
			if !_contains(auth.Keys, credential) {
				return responseError(c, fiber.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key", false)
			}
		case "bearer":
			token := strings.TrimSpace(strings.TrimPrefix(credential, "Bearer"))
			if !_contains(auth.Keys, token) {
				return responseError(c, fiber.StatusUnauthorized, "INVALID_BEARER_TOKEN", "Invalid bearer token", false)
			}
		default:
			return responseError(c, fiber.StatusInternalServerError, "UNSUPPORTED_AUTH_TYPE", "Unsupported authentication type", false)
		}

		return c.Next()
	}
}

// contains checks if a string exists in a slice
func _contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
