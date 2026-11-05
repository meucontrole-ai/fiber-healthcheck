# Architecture

`fiber-healthcheck` is a Go library (no binary, no deployment) that implements the MeuControle HTTP health contract described in the [README](../README.md). Services import it; it never imports a service.

## Layers

| Package | Role |
| --- | --- |
| `pkg/domain` | Contract types: `HealthStatus` (`Healthy`, `Degraded`, `Unhealthy`), `HealthCheckResult`, the `HealthCheck` / `ContextHealthCheck` interfaces and the `Healthy` / `Degraded` / `Unhealthy` result factories. No dependencies. |
| `pkg/usecase` | `HealthCheckManager`: registers checks, runs them, aggregates the status and maps it to an HTTP code. Also the naming standard (`ValidateName`, `AddPing`, `AddConfig`, `Severity`). Depends only on `domain`. |
| `pkg/adapter/fiber` | Fiber v3 handlers and `Register` (mounts `/health`, `/health/ready`, `/health/live` plus aliases) and `DrainOnShutdown`. |
| `pkg/adapter/nethttp` | The same routes on a standard `http.ServeMux`, for gRPC services and workers that expose an admin HTTP port. |
| `examples/basic` | Runnable Fiber example. |

Dependency direction is adapters → usecase → domain; adding a transport means adding an adapter, never touching the manager.

## Run lifecycle

1. A request reaches `/health` or `/health/ready`; the adapter calls `HealthCheckManager.Run(ctx)`.
2. If a run finished less than `CacheTTL` ago (default 1s) its result is returned; if one is in flight, the caller waits for it. Probe floods therefore cost one run.
3. Every check runs concurrently with its own `Timeout` (default 3s). A timeout, a panic or a `nil` result becomes `Unhealthy`.
4. Aggregation: any `Unhealthy` → `Unhealthy` (503); else any `Degraded` → `Degraded` (200); else `Healthy` (200). Entries are sorted by name.
5. `OnResult` receives each full result (for logging); the response hides `exception` and `data` unless `ExposeDetails` is true.

`/health/live` runs no check: it answers 200 while the process serves HTTP.

## Built-in `serving` check

Every manager starts with `serving`, `Healthy` until `Drain()` flips it to `Unhealthy` and clears the cache. The Fiber adapter drains on `app.Shutdown()`; other servers call `Drain()` before they stop accepting work. The names `serving` and `self` are reserved.
