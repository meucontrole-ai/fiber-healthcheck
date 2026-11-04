package main

import (
	"context"
	"log"
	"time"

	fiber_adapter "github.com/meucontrole-ai/fiber-healthcheck/pkg/adapter/fiber"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/domain"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/usecase"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	hc := usecase.NewHealthCheckManager(usecase.HealthCheckOptions{
		Timeout: 2 * time.Second,
		OnResult: func(name string, r domain.HealthCheckResult) {
			if r.Status != domain.HealthStatusHealthy && r.Exception != nil {
				log.Printf("health %s %s: %s", name, r.Status, *r.Exception)
			}
		},
	})

	hc.AddCheckFunc("self", func() *domain.HealthCheckResult {
		return domain.Healthy("process is up", nil)
	})
	hc.AddCheckContext("database", func(ctx context.Context) *domain.HealthCheckResult {
		// return pingDB(ctx) ...
		return domain.Healthy("database connection is ok", nil)
	})

	// Before any auth middleware.
	fiber_adapter.Register(app, hc)

	log.Fatal(app.Listen(":3000"))
}
