# fiber-healthcheck

Middleware for health check in [Fiber v3](https://github.com/gofiber/fiber).

## Installation

```bash
go get github.com/CarlosHe/fiber-healthcheck
```

## Sample Fiber HealthCheck

```go
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

	// 3. Register Handler
	app.Get("/healthcheck", fiber_adapter.NewHandler(hcManager))

	log.Fatal(app.Listen(":3000"))
}
```

#### HealthCheck json result

By calling the route `/healthcheck`, you will get the following structure:

```json
{
    "health_check": [
        {
            "ping": {
                "status": "Healthy",
                "description": "basic ping-pong health check",
                "exception": null,
                "data": {},
                "duration": "00:00:00.000"
            }
        },
        {
            "database": {
                "status": "Healthy",
                "description": "database connection is ok",
                "exception": null,
                "data": {
                    "db": "postgres"
                },
                "duration": "00:00:00.050"
            }
        }
    ]
}
```

## Other methods of use

### You can also create a struct and implement the HealthCheck interface

**Defining the struct in HealthCheckManager**

```go
hcManager.AddCheck("my-custom-check", &MyCustomCheck{})
```

**Implementation of the HealthCheck interface**

```go
type MyCustomCheck struct{}

func (c *MyCustomCheck) CheckHealth() *domain.HealthCheckResult {
	return domain.Healthy("custom check description", nil)
}
```

## Knowing the structure of HealthCheckResult

### Some ways to create the result, see some examples:

**Creating a Healthy result with description:**

```go
func (c *PingCheck) CheckHealth() *domain.HealthCheckResult {
    return domain.Healthy("description", nil)
}
```

**Creating a Unhealthy result with description and exception:**

```go
func (c *PingCheck) CheckHealth() *domain.HealthCheckResult {
    return domain.Unhealthy("description", errors.New("connection failed"), nil)
}
```

**Creating a Degraded result with description:**

```go
func (c *PingCheck) CheckHealth() *domain.HealthCheckResult {
    return domain.Degraded("description", nil, nil)
}
```

**Or using NewResult and passing the status with description:**

```go
func (c *PingCheck) CheckHealth() *domain.HealthCheckResult {
    return domain.NewResult(domain.HealthStatusHealthy, "description", nil, nil)
}
```

## License

MIT License
