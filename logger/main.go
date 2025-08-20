package logger

import (
	"fmt"
	"net/http"
	"time"
)

import (
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
)

// Returns the formatted server URL with cyan color for console output.
func GetServerHost(port string) string {
	serverUrlColor := color.New(color.FgCyan).SprintFunc()
	_host := "localhost"
	serverUrl := fmt.Sprintf("http://%s%s", _host, port)

	return serverUrlColor(serverUrl)
}

// Prints a standardized success message when the server starts.
func LogServerStart(port string) {
	LogSuccess(fmt.Sprintf("Server started on %s", GetServerHost(port)), 1)
}

// LogRoute logs detailed information about a single HTTP request.
// It includes method, path, IP, status code, response time, and optional prefix.
func LogRoute(method, path, ip string, status int, duration time.Duration, prefix string) {
	methodColors := map[string]*color.Color{
		"GET":     color.New(color.FgHiGreen),
		"POST":    color.New(color.FgHiCyan),
		"PUT":     color.New(color.FgYellow),
		"DELETE":  color.New(color.FgHiRed),
		"PATCH":   color.New(color.FgMagenta),
		"OPTIONS": color.New(color.FgHiWhite),
	}

	methodColor, ok := methodColors[method]
	if !ok {
		methodColor = color.New(color.FgWhite, color.Bold)
	}

	// Determine status color based on HTTP status code ranges
	var statusColor *color.Color
	switch {
	case status >= 500:
		statusColor = color.New(color.FgRed, color.Bold)
	case status >= 400:
		statusColor = color.New(color.FgHiYellow)
	case status >= 300:
		statusColor = color.New(color.FgYellow)
	case status >= 200:
		statusColor = color.New(color.FgGreen)
	default:
		statusColor = color.New(color.FgWhite)
	}

	ipColor := color.New(color.FgWhite)
	pathColor := color.New(color.FgHiBlack)
	durationColor := color.New(color.FgMagenta)
	prefixColor := color.New(color.FgHiBlue, color.Bold)

	prefixLog := ""
	if prefix != "" {
		prefixLog = prefixColor.Sprintf("%s", prefix)
	}

	// Compose log message
	msg := fmt.Sprintf(
		"%s %s %s",
		prefixLog,
		methodColor.Sprintf("%-7s", method),
		pathColor.Sprint(path),
	)

	if ip != "" {
		msg += " ip=" + ipColor.Sprint(ip)
	}

	if status > 0 {
		statusText := http.StatusText(status)
		msg += " " + statusColor.Sprintf("%d %s", status, statusText)
	}

	if duration > 0 {
		msg += " " + durationColor.Sprintf("%.2fms", float64(duration.Milliseconds()))
	}

	fmt.Println(msg)
}

// RequestLogger is a Fiber middleware that logs incoming HTTP requests.
// It captures method, path, status code, duration, and optionally client IP.
// Log level and color are determined based on the response status code.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next() // call next middleware/handler
		duration := time.Since(start)

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		msg := fmt.Sprintf("%s %s (%d) %dms", method, path, status, duration.Milliseconds())

		// Log based on status severity
		switch {
		case status >= 500:
			LogError(msg)
		case status >= 400:
			LogWarn(msg)
		default:
			LogSuccess(msg)
		}
		return err
	}
}

// --- Log Helpers --- //
//
// This section provides standardized logging utilities.
// All helpers delegate to `logWithType`, which handles consistent formatting and colorization.
// - LogSuccess → prints success messages (green).
// - LogError   → prints error messages (red).
// - LogWarn    → prints warning messages (yellow).
// - LogInfo    → prints informational messages (blue).
//
// Each function accepts the log message and optional empty line padding (addEmptyLines).
// Designed to keep console output clean, color-coded, and developer-friendly.

func LogSuccess(msg string, addEmptyLines ...int) {
	logWithType("OK", successStyle, msg, addEmptyLines...)
}

func LogError(msg string, addEmptyLines ...int) {
	logWithType("ERROR", errorStyle, msg, addEmptyLines...)
}

func LogWarn(msg string, addEmptyLines ...int) {
	logWithType("WARN", warnStyle, msg, addEmptyLines...)
}

func LogInfo(msg string, addEmptyLines ...int) {
	logWithType("INFO", infoStyle, msg, addEmptyLines...)
}
