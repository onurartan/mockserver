package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
)

// PathNormalizerMiddleware sanitizes the request URL by removing duplicate slashes.
// This ensures that routes like "//console//dashboard" are treated as "/console/dashboard",
// preventing routing mismatches and improving SEO/canonical URL handling.
func PathNormalizerMiddleware(consolePath string) fiber.Handler {

	slashRegex := regexp.MustCompile(`/{2,}`)
	return func(c *fiber.Ctx) error {
		path := c.Path()
		originalPath := path

		if strings.Contains(path, "//") {
			path = slashRegex.ReplaceAllString(path, "/")
		}

		if path != originalPath {
			return c.Redirect(path, fiber.StatusMovedPermanently)
		}

		return c.Next()
	}
}

// RegisterFallback returns a Catch-All handler (404 Not Found).
// It should be registered as the last handler in the stack to trap unmatched requests.
func RegisterFallback() fiber.Handler {
	return func(c *fiber.Ctx) error {

		path := c.Path()
		method := c.Method()

		errorMessage := fmt.Sprintf(
			"The requested resource was not found: %s [%s]. Please check the endpoint path and HTTP method.",
			path,
			method,
		)

		return responseError(
			c,
			fiber.StatusNotFound,
			"ROUTE_NOT_FOUND",
			errorMessage,
			false,
		)
	}
}

// authMiddleware enforces access control based on the configuration.
// It prioritizes Route-Level authentication over Global authentication.
// Supports: API Key (Header/Query) and Bearer Token schemes.
func authMiddleware(globalAuth, routeAuth *msconfig.AuthConfig) fiber.Handler {

	// Determine effective configuration (Route > Global)
	authConf := globalAuth
	if routeAuth != nil {
		authConf = routeAuth
	}

	if authConf == nil || !authConf.Enabled {
		return func(c *fiber.Ctx) error { return c.Next() }
	}

	return func(c *fiber.Ctx) error {

		authType := strings.ToLower(authConf.Type)
		authIn := strings.ToLower(authConf.In)
		authName := authConf.Name

		// Configuration Sanity Check
		if authType == "" {
			return responseError(c, fiber.StatusInternalServerError, "AUTH_MISCONFIGURED", "Authentication type is missing", false)
		}

		var credential string
		switch strings.ToLower(authIn) {
		case "header":
			credential = c.Get(authName)
		case "query":
			credential = c.Query(authName)
		}

		if credential == "" && authType == "bearer" {
			credential = c.Get("Authorization")
		}

		if credential == "" {
			return responseError(c, fiber.StatusUnauthorized, "MISSING_CREDENTIAL", "Missing authentication credential", false)
		}

		// Validate Credential Scheme
		switch strings.ToLower(authConf.Type) {
		case "apikey":
			if !_contains(authConf.Keys, credential) {
				return responseError(c, fiber.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key", false)
			}
		case "bearer":

			token := credential

			if len(credential) > 7 && strings.EqualFold(credential[0:7], "Bearer ") {
				token = credential[7:]
			}

			// Bearer token extraction and validation
			token = strings.TrimSpace(token)

			if !_contains(authConf.Keys, token) {
				return responseError(c, fiber.StatusUnauthorized, "INVALID_BEARER_TOKEN", "Invalid bearer token", false)
			}
		default:
			return responseError(c, fiber.StatusInternalServerError, "UNSUPPORTED_AUTH_TYPE", "Unsupported authentication type", false)
		}

		return c.Next()
	}
}

// containsString is a helper to check for string existence in a slice.
func _contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
