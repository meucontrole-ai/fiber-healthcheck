// Package fiber_adapter provides the Fiber v3 adapter for the health check system.
// It exposes a handler that can be registered with a Fiber application to serve
// health check results over HTTP.
package fiber_adapter

import (
	"github.com/CarlosHe/fiber-healthcheck/pkg/usecase"

	"github.com/gofiber/fiber/v3"
)

// NewHandler creates a Fiber handler that uses the provided HealthCheckManager.
// It executes the health checks and returns a JSON response with the appropriate HTTP status code.
//
// Usage:
//
//	app.Get("/healthcheck", fiber_adapter.NewHandler(manager))
func NewHandler(m *usecase.HealthCheckManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		status, res := m.GetResult()
		return c.Status(status).JSON(res)
	}
}
