package nethttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meucontrole-ai/fiber-healthcheck/pkg/domain"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/usecase"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	m := usecase.NewHealthCheckManager(usecase.HealthCheckOptions{CacheTTL: -1})
	m.AddCheckFunc("postgres", func() *domain.HealthCheckResult { return domain.Healthy("ok", nil) })
	Register(mux, m)
	for path, want := range map[string]int{"/health": 200, "/health/live": 200, "/health/ready": 200} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != want || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: %d", path, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/health", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST: %d", rec.Code)
	}
}
