package server

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
	appinfo "mockserver/pkg/appinfo"
)

var mockCache sync.Map

// loadMockFile reads & caches JSON mock files
func loadMockFile(path string) (interface{}, error) {
	if cached, ok := mockCache.Load(path); ok {
		return cached, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	mockCache.Store(path, parsed)
	return parsed, nil
}

// setMap safely sets a key-value pair in a map if the map is non-nil.
func setMap(m map[string]interface{}, key string, value interface{}) {
	if m != nil {
		m[key] = value
	}
}

func replacePathParams(path string) string {
	return regexp.MustCompile(`:([a-zA-Z0-9_]+)`).ReplaceAllString(path, "{$1}")
}

func buildParameters(route msconfig.RouteConfig) []map[string]interface{} {
	var params []map[string]interface{}

	for name, def := range route.PathParams {
		params = append(params, map[string]interface{}{
			"name":        name,
			"in":          "path",
			"required":    true,
			"schema":      map[string]interface{}{"type": def.Type},
			"description": def.Description,
			"example":     def.Example,
		})
	}

	for name, def := range route.Query {
		params = append(params, map[string]interface{}{
			"name":        name,
			"in":          "query",
			"required":    def.Required,
			"schema":      map[string]interface{}{"type": def.Type},
			"description": def.Description,
			"example":     def.Example,
		})
	}

	for name, def := range route.RequestHeaders {
		params = append(params, map[string]interface{}{
			"name":        name,
			"in":          "header",
			"required":    def.Required,
			"schema":      map[string]interface{}{"type": def.Type},
			"description": def.Description,
			"example":     def.Example,
		})
	}

	return params
}

func buildRequestBody(route msconfig.RouteConfig) map[string]interface{} {
	reqBody := map[string]interface{}{
		"required": true,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": route.BodySchema,
			},
		},
	}
	if route.BodyExample != nil {
		reqBody["content"].(map[string]interface{})["application/json"].(map[string]interface{})["example"] = route.BodyExample
	}
	return reqBody
}

func buildResponses(route msconfig.RouteConfig) map[string]interface{} {
	responses := map[string]interface{}{}

	// CASE responses
	for _, cs := range route.Cases {
		statusCode := fmt.Sprintf("%d", cs.Then.Status)
		responses[statusCode] = map[string]interface{}{
			"description": fmt.Sprintf("Case response for condition: %s", cs.When),
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"example": cs.Then.Body,
				},
			},
		}
	}

	// Default response
	if route.Default != nil {
		statusCode := fmt.Sprintf("%d", route.Default.Status)
		responses[statusCode] = map[string]interface{}{
			"description": "Default response if no case matches",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"example": route.Default.Body,
				},
			},
		}
	}

	// STATEFUL RESPONSE HANDLING
	if route.Stateful != nil {
		action := route.Stateful.Action

		switch action {
		case "list":
			responses["200"] = jsonResponseExample("List items", []interface{}{})

		case "create":
			responses["201"] = jsonResponseExample("Item created", map[string]interface{}{})

		case "get":
			responses["200"] = jsonResponseExample("Item found", map[string]interface{}{})
			responses["404"] = errorResponse("Not found", "Ensure the item exists or create it first")

		case "update":
			responses["200"] = jsonResponseExample("Item updated", map[string]interface{}{})
			responses["404"] = errorResponse("Not found", "Ensure the item exists before updating")

		case "delete":
			responses["200"] = jsonResponseExample("Item deleted", map[string]interface{}{
				"success": true,
			})
			responses["404"] = errorResponse("Not found", "Ensure the item exists before deleting")
		}

	}

	if route.Mock != nil && route.Mock.File != "" {
		if example, err := loadMockFile(route.Mock.File); err == nil {
			responses["200"] = map[string]interface{}{
				"description": "Successful response",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{"example": example},
				},
			}
		} else if route.Mock.Body != nil {
			responses["200"] = map[string]interface{}{
				"description": "Successful response",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{"example": route.Mock.Body},
				},
			}
		} else {
			responses["500"] = map[string]interface{}{
				"description": fmt.Sprintf("Failed to load mock file: %v", err),
			}
		}
	} else if route.Fetch != nil && route.Fetch.URL != "" {
		responses["200"] = map[string]interface{}{
			"description": "Successful response from upstream service",
		}
	}

	return responses
}

func jsonResponseExample(desc string, example interface{}) map[string]interface{} {
	return map[string]interface{}{
		"description": desc,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"example": example,
			},
		},
	}
}

func errorResponse(msg, hint string) map[string]interface{} {
	return map[string]interface{}{
		"description": msg,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"example": map[string]interface{}{
					"error": msg,
					"hint":  hint,
				},
			},
		},
	}
}

// applyAuthToOperation applies authentication metadata to an OpenAPI operation.
// Supports API key, bearer token, and basic auth.
func applyAuthToOperation(op map[string]interface{}, params *[]map[string]interface{}, auth *msconfig.AuthConfig) {
	if auth == nil || !auth.Enabled {
		return
	}

	var secName string
	switch strings.ToLower(auth.Type) {
	case "apikey":
		secName = "ApiKeyAuth"
		if auth.In != "" && auth.Name != "" {
			*params = append(*params, map[string]interface{}{
				"name":        auth.Name,
				"in":          auth.In,
				"required":    true,
				"schema":      map[string]interface{}{"type": "string"},
				"description": fmt.Sprintf("%s authentication token", strings.Title(auth.Type)),
			})
		}
	case "bearer":
		secName = "BearerAuth"
	case "basic":
		secName = "BasicAuth"
	}

	if secName != "" {
		setMap(op, "security", []map[string][]string{{secName: {}}})
	}
}

// [IMP_FUNC]
// generateOpenAPISpec generates an OpenAPI 3 spec from the mock server config.
// It handles tags, security schemes, parameters, request bodies, and responses.
func generateOpenAPISpec(cfg *msconfig.Config) map[string]interface{} {
	paths := make(map[string]interface{})
	var tags []map[string]string
	securitySchemes := make(map[string]interface{})

	// Tags
	for _, group := range cfg.Groups {
		tags = append(tags, map[string]string{
			"name":        group.Name,
			"description": group.Description,
		})
	}

	// Global Security Schemes
	if cfg.Server.Auth != nil && cfg.Server.Auth.Enabled {
		switch strings.ToLower(cfg.Server.Auth.Type) {
		case "apikey":
			securitySchemes["ApiKeyAuth"] = map[string]interface{}{
				"type": "apiKey", "in": cfg.Server.Auth.In, "name": cfg.Server.Auth.Name,
			}
		case "bearer":
			securitySchemes["BearerAuth"] = map[string]interface{}{
				"type": "http", "scheme": "bearer", "bearerFormat": "JWT",
			}
		case "basic":
			securitySchemes["BasicAuth"] = map[string]interface{}{
				"type": "http", "scheme": "basic",
			}
		}
	}

	// Routes
	for _, route := range cfg.Routes {
		fullPath := cfg.Server.APIPrefix + replacePathParams(route.Path)
		method := strings.ToLower(route.Method)

		var description string

		if route.Description != "" {
			description = route.Description
		} else {
			description = fmt.Sprintf("Auto-generated route for %s", route.Name)
		}

		operation := map[string]interface{}{
			"summary":     route.Name,
			"description": description,
			"responses":   map[string]interface{}{},
		}

		if route.Tag != "" {
			operation["tags"] = []string{route.Tag}
		}

		parameters := buildParameters(route)

		// Auth
		auth := route.Auth
		if auth == nil {
			auth = cfg.Server.Auth
		}
		applyAuthToOperation(operation, &parameters, auth)

		if len(parameters) > 0 {
			operation["parameters"] = parameters
		}

		if route.BodySchema != nil {
			operation["requestBody"] = buildRequestBody(route)
		}

		operation["responses"] = buildResponses(route)

		// Add to paths
		if paths[fullPath] == nil {
			paths[fullPath] = make(map[string]interface{})
		}
		paths[fullPath].(map[string]interface{})[method] = operation
	}

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "MockServer API",
			"version": appinfo.Version,
		},
		"paths": paths,
	}
	if len(tags) > 0 {
		spec["tags"] = tags
	}
	if len(securitySchemes) > 0 {
		spec["components"] = map[string]interface{}{"securitySchemes": securitySchemes}
	}

	return spec
}

// [IMP_FUNC]
// swaggerUIHandler serves the Swagger UI for the API.
// Loads OpenAPI spec from /openapi.json endpoint.
func swaggerUIHandler(c *fiber.Ctx) error {
	const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8" />
<title>MockServer API Docs</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
window.onload = () => {
  SwaggerUIBundle({
    url: "/openapi.json",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis],
    layout: "BaseLayout",
    persistAuthorization: true
  })
}
</script>
</body>
</html>`
	return c.Type("html").SendString(swaggerHTML)
}
