// Package fiber_adapter exposes the health check manager on a Fiber v3 app.
package fiber_adapter

import (
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/domain"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/usecase"

	"github.com/gofiber/fiber/v3"
)

// Standard paths of the meucontrole health contract.
const (
	PathHealth = "/health"       // canonical: runs every check, 200 Healthy/Degraded, 503 Unhealthy
	PathLive   = "/health/live"  // process is up; never touches dependencies
	PathReady  = "/health/ready" // same as /health
)

func noStore(c fiber.Ctx) {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
}

// NewHandler serves the aggregated report of m.
func NewHandler(m *usecase.HealthCheckManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		code, report := m.Run(c.Context())
		noStore(c)
		return c.Status(code).JSON(report)
	}
}

// NewLivenessHandler answers 200 while the process can serve HTTP.
func NewLivenessHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		noStore(c)
		return c.Status(fiber.StatusOK).JSON(usecase.Report{Status: domain.HealthStatusHealthy, HealthCheck: []map[string]*domain.HealthCheckResult{}})
	}
}

// DrainOnShutdown drains every manager when app starts shutting down, so "serving" turns
// Unhealthy and probes route away before the listener closes.
func DrainOnShutdown(app *fiber.App, ms ...*usecase.HealthCheckManager) {
	app.Hooks().OnPreShutdown(func() error {
		for _, m := range ms {
			m.Drain()
		}
		return nil
	})
}

// Register mounts /health, /health/live and /health/ready plus legacy aliases answering like /health,
// before auth middleware; on a *fiber.App it also drains m on shutdown (see DrainOnShutdown).
func Register(r fiber.Router, m *usecase.HealthCheckManager, aliases ...string) {
	if app, ok := r.(*fiber.App); ok {
		DrainOnShutdown(app, m)
	}
	h := NewHandler(m)
	r.Get(PathHealth, h)
	r.Get(PathReady, h)
	r.Get(PathLive, NewLivenessHandler())
	for _, a := range aliases {
		r.Get(a, h)
	}
}
