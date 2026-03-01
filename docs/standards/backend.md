# Backend Standards

## Tech Choices
- Golang
  - Use github.com/benjamin-wright/games-hub as the root repo in module names
- NATS message passing for inter-service comms
- Microservices
  - common app framework
  - service providers implement client libraries that service consumers utilise
- Helm charts for deployment, shared charts where feasible for microservice apps
- K3D for local testing (the shared cluster defined in the root Makefile)
- Tilt for local test deployments
  - A Tiltfile deployment function for each application
  - A Tiltfile entrypoint for each application that calls the deployment function and also deploys / runs integration tests
  - A tiltfile in the root the calls the deployment functions of all the components and also deploys / runs the end to end tests

## Shared Tools Directory

Shared utilities live under `tools/` and must be used rather than reimplementing inline:

| Type | Location | Usage |
|------|----------|-------|
| Tilt utilities | `tools/tilt/utils.tiltfile` | `load("<relative-path>/tools/tilt/utils.tiltfile", "<symbol>")` |
| Go libraries | `tools/go/` | import as a module |
| Docker base images | `tools/docker/` | reference in Dockerfiles |

## Tool-Generated Files

Never directly create or edit files that are owned and managed by CLI tooling:

| File | Required command |
|------|------------------|
| `go.mod` | `go mod init <module>` |
| `go.sum` | `go get <dependency>` / `go mod tidy` |
| `go.mod` dependency versions | `go get <package>@latest` |

Do not approximate, guess, or hard-code version numbers in tool-generated files.

## External Dependency Ownership

All interaction with an external system (database, message broker, HTTP service) must be encapsulated in a single package behind an exported interface. Other packages depend on the interface, never on the external system directly. This ensures that consumers can be unit-tested with fakes and that external-system concerns (connection handling, transactions, retries) live in one place.

## Health Endpoints
- Every deployable service exposes a health endpoint for platform readiness and liveness checks.