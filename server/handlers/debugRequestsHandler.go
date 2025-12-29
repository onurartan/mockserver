package server_handlers

import (
	"strings"
	"time"

	"sync/atomic"
	"strconv"
	// "github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
)

import (
	msconfig "mockserver/config"
)

type RequestLog struct {
	ID         string    `json:"id"`
	Time       time.Time `json:"time"`
	DurationMs int64     `json:"duration_ms"`

	Request struct {
		Method string            `json:"method"`
		Path   string            `json:"path"`
		Query  map[string]string `json:"query,omitempty"`
		IP     string            `json:"ip"`
		UA     string            `json:"user_agent,omitempty"`
	} `json:"request"`

	Response struct {
		Status int `json:"status"`
	} `json:"response"`

	Route struct {
		Name string `json:"name,omitempty"`
		Type string `json:"type"` // mock | fetch | internal
	} `json:"route"`

	Upstream *struct {
		URL        string `json:"url"`
		Status     int    `json:"status"`
		DurationMs int64  `json:"duration_ms"`
	} `json:"upstream,omitempty"`
}

var (
	requestLogs   = make([]RequestLog, 0, 100)
	logChannel    = make(chan RequestLog, 2000)
	getLogsChan   = make(chan chan []RequestLog)
	maxLogRecords = 100
)

var IgnoredPaths = map[string]bool{
	"/openapi.json": true,
	"/favicon.ico":  true,
}

// goroutine
func StartLogAggregator() {

	go func() {
		for {
			select {
			case entry := <-logChannel:
				if len(requestLogs) >= maxLogRecords {
					requestLogs = requestLogs[1:]
				}
				requestLogs = append(requestLogs, entry)

			case respChan := <-getLogsChan:
				// Debug  logs filters
				filteredLogs := make([]RequestLog, 0, len(requestLogs))
				for _, log := range requestLogs {
					if log.Route.Type != "internal" && !IgnoredPaths[log.Request.Path] {
						filteredLogs = append(filteredLogs, log)
					}

				}
				respChan <- filteredLogs
			}
		}
	}()
}

// Utils
func extractSafeHeaders(c *fiber.Ctx) map[string]string {
	out := map[string]string{}

	if ua := c.Get("User-Agent"); ua != "" {
		out["user-agent"] = ua
	}
	if ct := c.Get("Content-Type"); ct != "" {
		out["content-type"] = ct
	}
	if al := c.Get("Accept-Language"); al != "" {
		out["accept-language"] = al
	}
	return out
}

func getClientIP(c *fiber.Ctx) string {
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return c.IP()
}

// func safeQueries(queries map[string]string) map[string]string {
// 	safeQueries := make(map[string]string, len(queries))
// 	for k, v := range queries {
// 		safeQueries[string([]byte(k))] = string([]byte(v))
// 	}
// 	return safeQueries
// }

var requestCounter uint64

// Middleware
func RequestLoggerMiddleware(debugPath string, cfg *msconfig.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if strings.HasPrefix(c.Path(), debugPath) || IgnoredPaths[c.Path()] || strings.HasPrefix(c.Path(), cfg.Server.Console.Path) {
			return c.Next()
		}

		start := time.Now()
		// reqID := uuid.NewString()

		count := atomic.AddUint64(&requestCounter, 1)
		reqID := strconv.FormatUint(count, 10)
		c.Locals(CtxRequestID, reqID)

		// SAFE SNAPSHOT (BEFORE Next)
		method := string([]byte(c.Method()))
		originalURL := string([]byte(c.OriginalURL()))
		// queries := safeQueries(c.Queries())
		queries := c.Queries()
		ip := getClientIP(c)
		ua := string([]byte(c.Get("User-Agent")))

		err := c.Next()

		entry := RequestLog{
			ID:         reqID,
			Time:       start,
			DurationMs: time.Since(start).Milliseconds(),
		}

		entry.Request.Method = method
		entry.Request.Path = originalURL
		entry.Request.Query = queries
		entry.Request.IP = ip
		entry.Request.UA = ua
		entry.Response.Status = c.Response().StatusCode()

		if v := c.Locals(CtxRouteType); v != nil {
			entry.Route.Type = v.(string)
		}
		if v := c.Locals(CtxRouteName); v != nil {
			entry.Route.Name = v.(string)
		}

		if v := c.Locals(CtxUpstreamURL); v != nil {
			entry.Upstream = &struct {
				URL        string `json:"url"`
				Status     int    `json:"status"`
				DurationMs int64  `json:"duration_ms"`
			}{
				URL:        v.(string),
				Status:     c.Locals(CtxUpstreamStatus).(int),
				DurationMs: c.Locals(CtxUpstreamTimeMs).(int64),
			}
		}

		select {
		case logChannel <- entry:
		default:
		}

		return err
	}
}

func DebugRequestsHandler(c *fiber.Ctx) error {
	respChan := make(chan []RequestLog)
	getLogsChan <- respChan
	logs := <-respChan

	return c.JSON(logs)
}
