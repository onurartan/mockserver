package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	
	"regexp"
	"strings"
	"time"
)

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gofiber/fiber/v2"
)

import (
	server_utils "mockserver/server/utils"
)

// validateDelay checks if the provided delay (in milliseconds) is valid.
// Ensures the delay does not exceed 10 seconds (10000 ms).
// Returns the valid delay or an error if the limit is exceeded.
func validateDelay(delay int) (int, error) {
	if delay > 10000 {
		return 0, fmt.Errorf("delay cannot exceed 10000 ms (10 seconds), got %d", delay)
	}
	return delay, nil
}

// mergeHeaders merges three sets of HTTP headers into one.
// Priority order: defaults < routeHeaders < customHeaders
// meaning later headers overwrite earlier ones if the same key exists.
func mergeHeaders(defaults, routeHeaders, customHeaders map[string]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range defaults {
		headers[k] = v
	}
	for k, v := range routeHeaders {
		headers[k] = v
	}
	for k, v := range customHeaders {
		headers[k] = v
	}
	return headers
}

// parseAndFilterMockData takes raw JSON template data, applies fake data generation
// (via processTemplateJSON), unmarshals it into []map[string]interface{} and
// applies query parameter filters using server_utils.FilteredMockData.
// Returns the filtered slice or an error.
func parseAndFilterMockData(data []byte, params map[string]string) ([]map[string]interface{}, error) {
	processed, err := processTemplateJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to process template JSON: %w", err)
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(processed, &arr); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	filtered, err := server_utils.FilteredMockData(arr, params)
	if err != nil {
		return nil, fmt.Errorf("failed to filter mock data: %w", err)
	}
	return filtered, nil
}

// buildTargetURL builds the final request URL for fetch proxying.
// - Replaces path parameters in the base URL (e.g. {id} → 123)
// - Appends/overwrites query parameters
func buildTargetURL(base *url.URL, pathParams, queryParams map[string]string) string {
	target := *base
	path := target.Path
	for k, v := range pathParams {
		path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", k), v)
	}
	target.Path = path

	q := target.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	target.RawQuery = q.Encode()
	return target.String()
}

// processTemplateJSON parses JSON byte slices and replaces placeholders
// like {{name}}, {{uuid}}, {{email}}, {{number min=10 max=99}} with
// dynamically generated fake values using gofakeit.
// Supports: name, uuid, email, bool, date, number(min/max).
func processTemplateJSON(data []byte) ([]byte, error) {
	re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)([^}]*)}}`)
	return re.ReplaceAllFunc(data, func(match []byte) []byte {
		full := string(match)
		parts := re.FindStringSubmatch(full)
		if len(parts) < 2 {
			return match
		}
		key := parts[1]
		args := strings.TrimSpace(parts[2])

		switch key {
		case "name":
			return []byte(gofakeit.Name())
		case "uuid":
			return []byte(gofakeit.UUID())
		case "email":
			return []byte(gofakeit.Email())
		case "bool":
			return []byte(fmt.Sprintf("%v", gofakeit.Bool()))
		case "date":
			return []byte(gofakeit.Date().Format("2006-01-02"))
		case "number":
			// number min=10 max=99
			min, max := 1, 1000
			if args != "" {
				for _, arg := range strings.Fields(args) {
					if strings.HasPrefix(arg, "min=") {
						fmt.Sscanf(arg, "min=%d", &min)
					}
					if strings.HasPrefix(arg, "max=") {
						fmt.Sscanf(arg, "max=%d", &max)
					}
				}
			}
			return []byte(fmt.Sprintf("%d", gofakeit.Number(min, max)))
		default:
			return match
		}
	}), nil
}

// responseError sends a standardized error response in JSON format.
// If returnObject=true, it both writes the JSON error to response
// and returns the ApiError struct as error (for handler usage).
// Otherwise, only writes the JSON error response and returns nil.
func responseError(c *fiber.Ctx, status int, errCode, message string, returnObject bool) error {
	apiErr := &ApiError{
		Success:   false,
		Status:    status,
		Err:       http.StatusText(status),
		ErrorCode: errCode,
		Message:   message,
		Timestamp: time.Now().UTC().UnixNano() / 1e6,
	}

	if returnObject == true {
		_ = c.Status(status).JSON(apiErr)

		return apiErr
	}

	return c.Status(status).JSON(apiErr)
}
