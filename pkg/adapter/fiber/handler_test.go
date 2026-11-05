package fiber_adapter

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/domain"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/usecase"
)

func TestRegister(t *testing.T) {
	app := fiber.New()
	m := usecase.NewHealthCheckManager(usecase.HealthCheckOptions{CacheTTL: -1})
	m.AddCheckFunc("db", func() *domain.HealthCheckResult { return domain.Unhealthy("down", nil, nil) })
	Register(app, m, "/healthz")
	for path, want := range map[string]int{"/health": 503, "/health/ready": 503, "/healthz": 503, "/health/live": 200} {
		res, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != want {
			t.Fatalf("%s: %d want %d", path, res.StatusCode, want)
		}
		if res.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: missing no-store", path)
		}
		var body usecase.Report
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil || body.Status == "" {
			t.Fatalf("%s: bad body %v %+v", path, err, body)
		}
	}
}

func TestRegisterDrainsOnShutdown(t *testing.T) {
	m := usecase.NewHealthCheckManager(usecase.HealthCheckOptions{CacheTTL: -1})
	app := fiber.New()
	Register(app, m)

	if err := app.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if !m.Draining() {
		t.Fatal("manager not drained on shutdown")
	}
}
