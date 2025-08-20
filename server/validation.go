package server

import (
	"fmt"
	"strconv"
	"strings"

	"net/http"
)

import (
	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
)

// validateRequestParams returns a Fiber middleware handler that validates incoming
// request parameters (path, query, and headers) against the route configuration.
//
// Validation rules are defined in msconfig.RouteConfig:
//   - Required parameters must be present, otherwise a 400 error is returned.
//   - Parameter values are type-checked (string, integer, boolean).
//   - Enum values are enforced if defined (e.g. status ∈ ["active","inactive"]).
//
// If a validation check fails, the function immediately sends a standardized
// JSON error response (via responseError) and stops further processing.
// Otherwise, the middleware passes control to the next handler.
//
// Example usage:
//
//	app.Get("/users/:id", validateRequestParams(routeConfig), handlerFunc)
//
// This ensures that ":id" and any query/header parameters respect the schema
// before `handlerFunc` is executed.
func validateRequestParams(route msconfig.RouteConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// [DEV_LOG] fmt.Printf("[Validator] Checking parameters for %s %s\n", route.Method, route.Path)

		// check is a small helper closure that validates a single parameter
		// against its definition (required, type, enum). If invalid, it sends
		// an error response and returns the error.
		check := func(raw, key string, def msconfig.ParamDef, kind string) error {
			if def.Required && raw == "" {
				// [DEV_LOG] fmt.Println("Response Error: Missing required", kind, key)
				resp_err := responseError(c,
					http.StatusBadRequest,
					fmt.Sprintf("MISSING_%s", strings.ToUpper(strings.ReplaceAll(kind, " ", "_"))),
					fmt.Sprintf("Missing required %s: %s", kind, key),
					true,
				)
				return resp_err
			}

			if raw != "" {
				if err := validateType(raw, def.Type); err != nil {
					//  [DEV_LOG] fmt.Println("Response Error: Invalid", kind, key, "-", err)
					resp_err := responseError(c,
						http.StatusBadRequest,
						fmt.Sprintf("INVALID_%s", strings.ToUpper(strings.ReplaceAll(kind, " ", "_"))),
						fmt.Sprintf("Invalid %s %s: %v", kind, key, err),
						true,
					)
					return resp_err
				}
				if err := validateEnum(raw, def.Enum); err != nil {
					// [DEV_LOG] fmt.Println("Response Error: Invalid enum value for", kind, key, "-", err)
					resp_err := responseError(c,
						http.StatusBadRequest,
						"INVALID_ENUM_VALUE",
						fmt.Sprintf("%s %s: %v", kind, key, err),
						true,
					)
					return resp_err
				}
			}

			return nil
		}

		// Path params
		for key, def := range route.PathParams {
			check_resp := check(c.Params(key), key, def, "path param")
			// [DEV_LOG] fmt.Println("Path param check:", key, "->", c.Params(key), "->", check_resp)
			if check_resp != nil && c.Response().StatusCode() != 0 {
				return check_resp
			}
		}

		// Query params
		for key, def := range route.Query {
			check_resp := check(c.Query(key), key, def, "query param")
			// [DEV_LOG] fmt.Println("Query param check:", key, "->", c.Params(key), "->", check_resp)
			if check_resp != nil && c.Response().StatusCode() != 0 {
				return check_resp
			}
		}

		// Headers
		for key, def := range route.RequestHeaders {
			check_resp := check(c.Get(key), key, def, "header")
			// [DEV_LOG]  fmt.Println("Header param check:", key, "->", c.Params(key), "->", check_resp)
			if check_resp != nil && c.Response().StatusCode() != 0 {
				return check_resp
			}
		}

		return nil
	}
}

// Checks raw string against type definition
func validateType(raw, typ string) error {
	switch strings.ToLower(typ) {
	case "string":
		return nil
	case "integer", "int":
		if _, err := strconv.Atoi(raw); err != nil {
			return fmt.Errorf("expected integer, got '%s'", raw)
		}
	case "boolean", "bool":
		if _, err := strconv.ParseBool(raw); err != nil {
			return fmt.Errorf("expected boolean, got '%s'", raw)
		}
	default:
		return fmt.Errorf("unsupported param type: %s", typ)
	}
	return nil
}

// Ensures raw value is one of the allowed enum values
func validateEnum(raw string, enum []string) error {
	if len(enum) == 0 {
		return nil
	}
	for _, v := range enum {
		if raw == v {
			return nil
		}
	}
	return fmt.Errorf("must be one of %v, got '%s'", enum, raw)
}
