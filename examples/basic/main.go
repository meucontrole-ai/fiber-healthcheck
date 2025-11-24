package main

import (
	"errors"
	"log"
	"time"

	fiber_adapter "github.com/CarlosHe/fiber-healthcheck/pkg/adapter/fiber"
	"github.com/CarlosHe/fiber-healthcheck/pkg/domain"
	"github.com/CarlosHe/fiber-healthcheck/pkg/usecase"

	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// 1. Create the Use Case
	hcManager := usecase.NewHealthCheckManager()

	// 2. Register Checks
	hcManager.AddCheckFunc("ping", func() *domain.HealthCheckResult {
		return domain.Healthy("basic ping-pong health check", nil)
	})

	hcManager.AddCheckFunc("database", func() *domain.HealthCheckResult {
		// Simulate DB check
		time.Sleep(50 * time.Millisecond)
		return domain.Healthy("database connection is ok", map[string]string{"db": "postgres"})
	})

	hcManager.AddCheckFunc("external-api", func() *domain.HealthCheckResult {
		// Simulate failure or degradation
		return domain.Degraded("api is slow", errors.New("latency high"), nil)
	})

	hcManager.AddCheckFunc("panic-check", func() *domain.HealthCheckResult {
		// Simulate panic or failure
		panic("unexpected error")
	})

	// 3. Register Handler
	app.Get("/healthcheck", fiber_adapter.NewHandler(hcManager))

	log.Fatal(app.Listen(":3000"))
}
