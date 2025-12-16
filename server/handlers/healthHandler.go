package server_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

import (
	appinfo "mockserver/internal/appinfo"
)

type HealthResponse struct {
	Status      string        `json:"status"`
	Uptime      string        `json:"uptime"`
	StartTime   time.Time     `json:"start_time"`
	RouteCount  int           `json:"route_count"`
	MockRoutes  int           `json:"mock_routes"`
	FetchRoutes int           `json:"fetch_routes"`
	Version     string        `json:"version"`
}

func HealthHandler(routeCount, mockCount, fetchCount int, version string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(HealthResponse{
			Status:      "ok",
			Uptime:      time.Since(appinfo.StartTime).String(),
			StartTime:   appinfo.StartTime,
			RouteCount:  routeCount,
			MockRoutes:  mockCount,
			FetchRoutes: fetchCount,
			Version:     version,
		})
	}
}
