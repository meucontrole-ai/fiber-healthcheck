// Package nethttp exposes the health check manager on a net/http mux, with the same
// paths and JSON schema as the Fiber adapter (for gRPC services and workers).
package nethttp

import (
	"encoding/json"
	"net/http"

	"github.com/meucontrole-ai/fiber-healthcheck/pkg/domain"
	"github.com/meucontrole-ai/fiber-healthcheck/pkg/usecase"
)

// Standard paths of the meucontrole health contract.
const (
	PathHealth = "/health"
	PathLive   = "/health/live"
	PathReady  = "/health/ready"
)

func write(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func getOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}
}

// NewHandler serves the aggregated report of m.
func NewHandler(m *usecase.HealthCheckManager) http.Handler {
	return getOnly(func(w http.ResponseWriter, r *http.Request) {
		code, report := m.Run(r.Context())
		write(w, code, report)
	})
}

// NewLivenessHandler answers 200 while the process can serve HTTP.
func NewLivenessHandler() http.Handler {
	return getOnly(func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, usecase.Report{Status: domain.HealthStatusHealthy, HealthCheck: []map[string]*domain.HealthCheckResult{}})
	})
}

// Register mounts /health, /health/live, /health/ready and legacy aliases on mux.
func Register(mux *http.ServeMux, m *usecase.HealthCheckManager, aliases ...string) {
	h := NewHandler(m)
	mux.Handle(PathHealth, h)
	mux.Handle(PathReady, h)
	mux.Handle(PathLive, NewLivenessHandler())
	for _, a := range aliases {
		mux.Handle(a, h)
	}
}
