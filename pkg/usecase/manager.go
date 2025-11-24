// Package usecase implements the business logic for the health check system.
// It provides the HealthCheckManager which orchestrates the execution of multiple
// health checks and aggregates their results.
package usecase

import (
	"fmt"
	"sync"
	"time"

	"github.com/CarlosHe/fiber-healthcheck/pkg/domain"
)

// HealthCheckManager manages the registration and execution of health checks.
// It is safe for concurrent use.
type HealthCheckManager struct {
	checks map[string]domain.HealthCheck
	mu     sync.RWMutex
}

// NewHealthCheckManager creates a new instance of HealthCheckManager.
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		checks: make(map[string]domain.HealthCheck),
	}
}

// AddCheck registers a new health check with the given name.
// If a check with the same name already exists, it will be overwritten.
func (m *HealthCheckManager) AddCheck(name string, check domain.HealthCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checks[name] = check
}

// funcCheck is a simple wrapper to allow functions to be used as HealthChecks.
type funcCheck struct {
	f func() *domain.HealthCheckResult
}

func (fc *funcCheck) CheckHealth() *domain.HealthCheckResult {
	return fc.f()
}

// AddCheckFunc registers a function as a health check.
// This is a convenience method for simple checks that don't require a full struct.
func (m *HealthCheckManager) AddCheckFunc(name string, f func() *domain.HealthCheckResult) {
	m.AddCheck(name, &funcCheck{f: f})
}

// GetResult executes all registered checks concurrently and returns the aggregated result.
// It returns the HTTP status code and the response body map.
//
// The status code is determined as follows:
//   - 503 Service Unavailable: If any check returns Unhealthy.
//   - 200 OK: If all checks are Healthy or Degraded.
func (m *HealthCheckManager) GetResult() (int, map[string]any) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]map[string]*domain.HealthCheckResult, 0, len(m.checks))

	var wg sync.WaitGroup

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	type checkResultContainer struct {
		Name   string
		Result *domain.HealthCheckResult
	}

	resultChan := make(chan checkResultContainer, len(m.checks))

	for name, check := range m.checks {
		wg.Add(1)
		go func(n string, c domain.HealthCheck) {
			defer wg.Done()

			start := time.Now()

			// Recover from panic to ensure the health check doesn't crash the application
			// and reports the failure correctly.
			defer func() {
				if r := recover(); r != nil {
					duration := time.Since(start)
					err := fmt.Errorf("panic: %v", r)
					// Create an Unhealthy result with the panic message
					res := domain.Unhealthy("Check panicked", err, nil)
					res.Duration = formatDuration(duration)
					resultChan <- checkResultContainer{Name: n, Result: res}
				}
			}()

			res := c.CheckHealth()
			duration := time.Since(start)

			res.Duration = formatDuration(duration)

			resultChan <- checkResultContainer{Name: n, Result: res}
		}(name, check)
	}

	wg.Wait()
	close(resultChan)

	for res := range resultChan {
		switch res.Result.Status {
		case domain.HealthStatusHealthy:
			healthyCount++
		case domain.HealthStatusDegraded:
			degradedCount++
		case domain.HealthStatusUnhealthy:
			unhealthyCount++
		}

		results = append(results, map[string]*domain.HealthCheckResult{
			res.Name: res.Result,
		})
	}

	statusCode := calculateStatusCode(healthyCount, degradedCount, unhealthyCount)

	response := map[string]any{
		"health_check": results,
	}

	return statusCode, response
}

func calculateStatusCode(healthy, degraded, unhealthy int) int {
	if unhealthy > 0 {
		return 503 // Service Unavailable
	}
	// In some implementations, Degraded might return 200 or something else.
	// Here we stick to 200 for Degraded as per the original Delphi logic reference.
	return 200
}

func formatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}
