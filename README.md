# fiber-healthcheck

The meucontrole health check contract for Go services: one JSON schema and one set of paths, served by Fiber v3 or `net/http`.

```bash
go get github.com/meucontrole-ai/fiber-healthcheck@latest
```

## The contract

| Path | Runs checks | Answer |
| --- | --- | --- |
| `GET /health` | yes | `200` Healthy or Degraded, `503` Unhealthy. **Canonical path**: probes, load balancers, monitors and agents.control point here. |
| `GET /health/ready` | yes | Same as `/health` (Kubernetes readiness). |
| `GET /health/live` | no | `200` while the process serves HTTP (Kubernetes liveness; never restarts a pod because a dependency is down). |

Body:

```json
{
  "status": "Healthy",
  "health_check": [
    { "database": { "status": "Healthy", "description": "postgres ping ok", "exception": null, "data": {}, "duration": "00:00:00.004" } }
  ]
}
```

Rules:

- Aggregate status: any `Unhealthy` makes it `Unhealthy` (503); otherwise any `Degraded` makes it `Degraded` (200); no checks is `Healthy`.
- Every check has a timeout (default 3s). A hung check is `Unhealthy`, the endpoint never hangs.
- Results are cached for `CacheTTL` (default 1s) and concurrent callers share one run, so probe floods never hit the database N times.
- `exception` and `data` are hidden unless `ExposeDetails: true`: error strings leak hosts, users and topology. Use `OnResult` to log them.
- Panics and `nil` results become `Unhealthy`; entries are sorted by name; responses carry `Cache-Control: no-store`.
- Mount the routes **before** auth middleware: a probe must never get 401.

## Fiber

```go
hc := usecase.NewHealthCheckManager(usecase.HealthCheckOptions{
	OnResult: func(name string, r domain.HealthCheckResult) {
		if r.Status != domain.HealthStatusHealthy && r.Exception != nil {
			log.Printf("health %s %s: %s", name, r.Status, *r.Exception)
		}
	},
})
hc.AddCheckContext("database", func(ctx context.Context) *domain.HealthCheckResult {
	if err := db.PingContext(ctx); err != nil {
		return domain.Unhealthy("postgres ping failed", err, nil)
	}
	return domain.Healthy("postgres ping ok", nil)
})

fiber_adapter.Register(app, hc)              // /health, /health/live, /health/ready
fiber_adapter.Register(app, hc, "/healthz")  // plus legacy aliases kept for existing probes
```

## net/http (gRPC services, workers)

Serve the same contract on the service's HTTP/admin port:

```go
mux := http.NewServeMux()
nethttp.Register(mux, hc)
```

## Checks

- `AddCheckContext(name, func(ctx) *domain.HealthCheckResult)`: preferred for anything doing I/O.
- `AddCheckFunc(name, func() *domain.HealthCheckResult)`, `AddCheck(name, domain.HealthCheck)`: a type that also implements `domain.ContextHealthCheck` gets the context.
- Results: `domain.Healthy(desc, data)`, `domain.Degraded(desc, err, data)`, `domain.Unhealthy(desc, err, data)`.

Check what the service needs to serve traffic (database, cache, broker, required upstream gRPC). Do not make an optional dependency `Unhealthy`; use `Degraded`.

## Migrating from v1.0.0 (`github.com/CarlosHe/fiber-healthcheck`)

1. Replace the import path with `github.com/meucontrole-ai/fiber-healthcheck`.
2. Replace `app.Get("/healthcheck", fiber_adapter.NewHandler(m))` with `fiber_adapter.Register(app, m, "/healthcheck")`.
3. The body gained a top-level `status`. A check with some `Unhealthy` result now answers 503 (v1.0.0 answered 200 when another check was healthy).

## License

MIT

## Built-in `serving` check

Every `HealthCheckManager` is born with a `serving` check: `Healthy` until `Drain()` is called,
`Unhealthy` afterwards. It is the only "the process is up" signal; never register your own `self`
or `serving` (both names are reserved and `AddCheck*` panics on them). Call `Drain()` when shutdown
starts, before the server stops accepting work, so probes and callers route away first. The Fiber
adapter `Register` does it on `app.Shutdown()`; `DrainOnShutdown(app, managers...)` covers extra
managers. A gRPC service calls `hc.Drain()` right before `GracefulStop()`.
