# Development and release

## Run locally

```bash
go vet ./...
go test -race -count=1 ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run ./examples/basic
```

The example serves the contract on Fiber; `curl -i localhost:3000/health` shows the aggregated body.

## CI

`.github/workflows/ci.yml` runs vet, race tests and govulncheck on the latest Go 1.26 patch on pushes to `main` and on pull requests.

## Dependency updates

`.github/dependabot.yml` opens weekly (Monday 06:00 America/Sao_Paulo) grouped minor/patch updates for Go modules and GitHub Actions against `main`.

## Release

Consumers pin semver tags (`v1.0.0` … `v1.3.0`). To release, merge to `main` with CI green and push an annotated tag `vX.Y.Z`; the Go module proxy picks it up. A breaking change to the JSON body, the paths or the public Go API needs a major version and a migration note in the README, like the one for v1.0.0.

## Adopting in a service

- HTTP services: `fiber_adapter.Register(app, hc)` before any auth middleware.
- gRPC services and workers that already expose HTTP: `nethttp.Register(mux, hc)`. Pure gRPC cores use `grpc-healthcheck` (`grpc.health.v1`) instead and expose no HTTP health.
- Register checks with `AddPing` / `AddConfig` and the standard names; optional dependencies are `Optional` (Degraded), never Unhealthy.
