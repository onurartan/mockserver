package server

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	msconfig "mockserver/config"
)

const (
	MaskedValue    = "********"
	JWTCookieName  = "ms_console_jwt"
	ContextUserKey = "user_claims" // Key used to store user claims in Fiber context
)

var jwtSecret []byte

// initJWTSecret initializes the JWT signing key.
// It prioritizes the environment variable; otherwise, it derives a deterministic key
// from the admin password to invalidate sessions upon password change.
func initJWTSecret(cfg *msconfig.Config) {
	if secret := os.Getenv("MS_JWT_SECRET"); secret != "" {
		jwtSecret = []byte(secret)
		return
	}
	jwtSecret = []byte(cfg.Server.Console.Auth.Password + "_ms_secure_salt_v1")
}

type ConsoleClaims struct {
	Username string `json:"u"`
	jwt.RegisteredClaims
}

// generateToken creates a signed JWT for the authenticated user.
// Tokens are valid for 72 hours.
func generateToken(username string) (string, error) {
	claims := ConsoleClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mockserver-console",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateToken(tokenString string) (*ConsoleClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ConsoleClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Enforce HMAC signing method to prevent "none" algorithm attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*ConsoleClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// ConsoleAuthMiddleware enforces stateless JWT authentication.
// It handles context differentiation (API vs Browser) and validates session consistency against the current config.
func ConsoleAuthMiddleware(cfg *msconfig.Config) fiber.Handler {
	// initJWTSecret(cfg) [OLD]

	return func(c *fiber.Ctx) error {
		path := c.Path()
		consolePath := cfg.Server.Console.Path

		// Bypass authentication for specific routes (Login, Public assets)
		if !cfg.Server.Console.Auth.Enabled ||
			strings.HasPrefix(path, consolePath+"/login") ||
			strings.HasPrefix(path, "/public") {
			return c.Next()
		}

		// Token Extraction & Validation
		tokenString := c.Cookies(JWTCookieName)
		claims, err := validateToken(tokenString)

		// handleAuthError determines the appropriate response format (JSON for XHR/API, Redirect for Browser).
		handleAuthError := func() error {
			c.ClearCookie(JWTCookieName)
			isAPI := strings.Contains(c.Get("Accept"), "application/json") ||
				c.XHR() ||
				strings.HasSuffix(path, ".json") ||
				strings.Contains(path, "/me")

			if isAPI {
				return c.Status(401).JSON(fiber.Map{
					"error": "Unauthorized Access",
					"code":  "AUTH_REQUIRED",
				})
			}
			return c.Redirect(consolePath + "/login")
		}

		if err != nil {
			return handleAuthError()
		}

		// Consistency Check (Stale Data Protection)
		// Even if the token signature is valid, we must ensure the user in the payload
		// matches the currently configured admin user.

		validUser := os.Getenv("MS_CONSOLE_USER")
		if validUser == "" {
			validUser = cfg.Server.Console.Auth.Username
		}

		if claims.Username != validUser {
			return handleAuthError()
		}

		// Store claims in context for downstream handlers
		c.Locals(ContextUserKey, claims)
		return c.Next()
	}
}

// ConsoleLoginHandler processes authentication credentials.
// It implements timing-attack safe comparisons and sets an HTTP-Only cookie upon success.
func ConsoleLoginHandler(cfg *msconfig.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&creds); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Malformed request"})
		}

		// Resolve credentials (Env override > Config)
		validUser := os.Getenv("MS_CONSOLE_USER")
		if validUser == "" {
			validUser = cfg.Server.Console.Auth.Username
		}

		validPass := os.Getenv("MS_CONSOLE_PASS")
		if validPass == "" {
			validPass = cfg.Server.Console.Auth.Password
		}

		// Prevent Timing Attacks using ConstantTimeCompare
		userMatch := subtle.ConstantTimeCompare([]byte(creds.Username), []byte(validUser)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(creds.Password), []byte(validPass)) == 1

		if userMatch && passMatch {
			signedToken, err := generateToken(creds.Username)
			if err != nil {
				return c.Status(500).SendString("Token error")
			}

			c.Cookie(&fiber.Cookie{
				Name:     JWTCookieName,
				Value:    signedToken,
				Expires:  time.Now().Add(72 * time.Hour),
				HTTPOnly: true, // Mitigate XSS
				Secure:   false,
				SameSite: "Lax", // CSRF Protection
			})

			return c.JSON(fiber.Map{"success": true, "redirect": cfg.Server.Console.Path})
		}

		time.Sleep(300 * time.Millisecond)
		return c.Status(401).JSON(fiber.Map{"success": false, "error": "Invalid credentials"})
	}
}

// ConsoleMeHandler returns the authenticated user's profile and UI preferences.
func ConsoleMeHandler(c *fiber.Ctx) error {
	claims, ok := c.Locals(ContextUserKey).(*ConsoleClaims)
	if !ok || claims == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Session expired"})
	}
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"username": claims.Username,
			"role":     "admin",
			"email":    claims.Username + "@mockserver.local",
			"avatar":   "https://avatar.roticeh.com/avatar/" + claims.Username + "?initials=auto&aType=color",
		},
	})
}

// ConsoleLogoutHandler invalidates the session and clears client-side cache
// to prevent "Back button" access to protected pages.
func ConsoleLogoutHandler(cfg *msconfig.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Invalidate Cookie (Set expiration to the past)
		c.Cookie(&fiber.Cookie{
			Name:     JWTCookieName, //
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
		})

		// Clear Browser Cache (Security Requirement)
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		// Handle response based on client type (API vs Browser)
		loginPath := cfg.Server.Console.Path + "/login" //

		if c.XHR() || strings.Contains(c.Get("Accept"), "application/json") {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"success":  true,
				"message":  "Logged out successfully",
				"redirect": loginPath,
			})
		}

		return c.Redirect(loginPath)
	}
}

// SafeConfigHandler returns a sanitized version of the server configuration
// for frontend consumption, masking sensitive secrets.
func SafeConfigHandler(cfg *msconfig.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Deep copy via serialization to avoid modifying the runtime config
		rawBytes, _ := json.Marshal(cfg)
		var safeCfg msconfig.Config
		json.Unmarshal(rawBytes, &safeCfg)

		// Mask Global Auth
		if safeCfg.Server.Auth != nil && safeCfg.Server.Auth.Enabled {
			safeCfg.Server.Auth.Keys = []string{MaskedValue}
		}

		// Mask Console Password
		if safeCfg.Server.Console != nil && safeCfg.Server.Console.Auth != nil {
			safeCfg.Server.Console.Auth.Password = MaskedValue
		}

		// Mask Route-specific Auth
		for i := range safeCfg.Routes {
			if safeCfg.Routes[i].Auth != nil && safeCfg.Routes[i].Auth.Enabled {
				safeCfg.Routes[i].Auth.Keys = []string{MaskedValue}
			}
		}

		return c.JSON(safeCfg)
	}
}

// ConsoleAssetGuard middleware protects static assets from hotlinking.
// It ensures that assets (.js, .css, .map) are only loaded within the console context.
func ConsoleAssetGuard(consoleCfg *msconfig.ConsoleConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Block bot traffic / requests without UA
		if c.Get("User-Agent") == "" {
			return fiber.ErrForbidden
		}

		path := c.Path()

		if strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".map") {

			ref := c.Get("Referer")

			// Reject if direct access (empty referer) or external linking
			if ref == "" || !strings.Contains(ref, consoleCfg.Path) {
				return fiber.ErrForbidden
			}
		}

		return c.Next()
	}
}
