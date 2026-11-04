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

// Register mounts /health, /health/live and /health/ready, plus any legacy alias paths
// (e.g. "/healthz", "/healthcheck") that answer like /health. Register it before
// authentication middleware so probes never get 401.
func Register(r fiber.Router, m *usecase.HealthCheckManager, aliases ...string) {
	h := NewHandler(m)
	r.Get(PathHealth, h)
	r.Get(PathReady, h)
	r.Get(PathLive, NewLivenessHandler())
	for _, a := range aliases {
		r.Get(a, h)
	}
}
