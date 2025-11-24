// Package domain defines the core entities and interfaces for the health check system.
// It contains the data structures representing health status and results, as well as
// the contract for implementing health checks.
package domain

// HealthStatus represents the status of a health check.
// It can be Healthy, Degraded, or Unhealthy.
type HealthStatus string

const (
	// HealthStatusHealthy indicates that the component is working correctly.
	HealthStatusHealthy HealthStatus = "Healthy"
	// HealthStatusDegraded indicates that the component is working but with some issues (e.g., high latency).
	HealthStatusDegraded HealthStatus = "Degraded"
	// HealthStatusUnhealthy indicates that the component is not working or is unreachable.
	HealthStatusUnhealthy HealthStatus = "Unhealthy"
)

// HealthCheckResult represents the detailed result of a single health check execution.
// It includes the status, a description, optional exception details, custom data, and execution duration.
type HealthCheckResult struct {
	// Status is the health status of the check (Healthy, Degraded, Unhealthy).
	Status HealthStatus `json:"status"`
	// Description provides a human-readable explanation of the check result.
	Description string `json:"description"`
	// Exception contains the error message if the check failed or is degraded.
	// It is a pointer to allow null values in JSON output.
	Exception *string `json:"exception"`
	// Data holds any additional context or metrics related to the check.
	Data any `json:"data"`
	// Duration indicates how long the check took to execute, formatted as a string.
	Duration string `json:"duration"`
}

// HealthCheck is the interface that must be implemented by any component
// that wants to provide a health check.
//
// Implementations should return a *HealthCheckResult indicating the current state of the component.
type HealthCheck interface {
	// CheckHealth executes the health check logic and returns the result.
	CheckHealth() *HealthCheckResult
}
