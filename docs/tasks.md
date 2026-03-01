# Tasks

An itemised list of the next most important changes to make, ordered by priority.

| # | Title | Priority | Status |
|---|-------|----------|--------|
| 8 | [DB Migrations — Go runner and base Docker image](#8-db-migrations--go-runner-and-base-docker-image) | High | Todo |
| 9 | [DB Migrations — Common Helm chart](#9-db-migrations--common-helm-chart) | High | Todo |
| 10 | [DB Migrations — Reusable Tilt functions](#10-db-migrations--reusable-tilt-functions) | High | Todo |

---

## 8: DB Migrations — Go runner and base Docker image

A standalone Go CLI that executes versioned SQL migrations against a PostgreSQL database, tracks applied state in a `_migrations` table with content hashes, and errors on tampered files. Packaged as a reusable base Docker image that apps extend by copying in their migration SQL files.

**Scope**
- `apps/platform/db-migrations/cmd/main.go`
  - CLI entrypoint; reads Postgres connection details from environment variables (`PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`) and accepts `--target` (optional migration ID) and `--migrations-dir` (default `/migrations`) flags
  - Exits 0 on success, non-zero on any error
- `apps/platform/db-migrations/internal/discovery/discovery.go`
  - Discovers migration files from the migrations directory
  - Parses `<id>-<name>-apply.sql` / `<id>-<name>-rollback.sql` pairs; errors on missing pairs or malformed filenames
  - Returns migrations sorted by ID
- `apps/platform/db-migrations/internal/discovery/discovery_test.go`
  - Unit tests: correct parsing, error on missing pair, correct ID ordering, rejection of malformed filenames
- `apps/platform/db-migrations/internal/store/store.go`
  - Ensures `_migrations` tracking table exists (`id TEXT PRIMARY KEY, name TEXT, applied_at TIMESTAMPTZ, apply_hash TEXT, rollback_hash TEXT`)
  - CRUD operations for migration records; stores and compares SHA-256 content hashes
- `apps/platform/db-migrations/internal/store/store_test.go`
  - Integration tests against a real Postgres instance: create tracking table, insert/query/delete records, hash comparison
- `apps/platform/db-migrations/internal/runner/runner.go`
  - Core execution logic: validates integrity of already-applied migrations (hash comparison), determines direction (forward apply / rollback / no-op) based on current state and optional target, executes each SQL file within a transaction, updates the tracking table
  - Uses an interface for the store to support unit testing with a fake
- `apps/platform/db-migrations/internal/runner/runner_test.go`
  - Unit tests: direction determination, hash mismatch detection, correct ordering of operations, no-op when fully applied
- `apps/platform/db-migrations/Makefile`
  - `build`, `test`, `integration-test`, `fmt`, `vet`, `lint`, `clean` targets (mirrors db-operator pattern)
- `tools/docker/migrations.Dockerfile`
  - Multi-stage build: builder compiles Go binary from `apps/platform/db-migrations/`, runtime stage uses `distroless/static:nonroot`, copies binary to `/usr/local/bin/migrate`
  - `ENTRYPOINT ["/usr/local/bin/migrate"]`
- `apps/platform/db-migrations/spec.md`
  - Add note that the runner is implemented in Go

**Acceptance Criteria**
- [ ] Runner discovers `<id>-<name>-apply.sql` / `<id>-<name>-rollback.sql` files and applies them in ID order
- [ ] `_migrations` tracking table records each applied migration with SHA-256 content hashes of apply and rollback files
- [ ] Runner errors if a previously-applied migration file's content hash has changed
- [ ] When `--target` is set ahead of current state, applies up to and including that ID
- [ ] When `--target` is set behind current state, rolls back from current down to (but not including) that ID in reverse order
- [ ] When no `--target` is specified, applies all unapplied migrations
- [ ] Base Docker image builds and is extensible (apps add migration files via `COPY` into `/migrations`)
- [ ] Unit tests pass for discovery, runner direction logic, and hash validation
- [ ] Integration tests pass for store operations against a real Postgres instance

**Dependencies:** None

**New dependencies:** `github.com/lib/pq` (PostgreSQL driver — already used in db-operator, added as direct dependency in the new module). No other external dependencies; uses stdlib for file I/O, `crypto/sha256`, sorting, and flag parsing.

---

## 9: DB Migrations — Common Helm chart

A reusable Helm chart that deploys a Kubernetes Job for running migrations. The Job name includes an incrementing index to avoid immutable-object patch conflicts on re-deploy. Accepts a target migration ID value for selective rollouts and rollbacks.

**Scope**
- `apps/platform/db-migrations/helm/Chart.yaml`
  - Chart metadata (`name: db-migrations`, `type: application`)
- `apps/platform/db-migrations/helm/values.yaml`
  - `image.repository`, `image.tag`, `image.pullPolicy`
  - `releaseIndex` (integer, incremented by caller on each deploy)
  - `targetMigration` (string, optional — omitted or empty means "apply all")
  - `db.secretName` (name of the K8s Secret containing `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`)
- `apps/platform/db-migrations/helm/templates/_helpers.tpl`
  - Standard helpers (fullname, labels, selector labels)
- `apps/platform/db-migrations/helm/templates/job.yaml`
  - Job name incorporates `releaseIndex` (e.g. `{{ fullname }}-{{ .Values.releaseIndex }}`) so each deploy creates a new Job
  - Container runs the migration image with env vars sourced from `db.secretName`
  - Conditionally passes `--target={{ .Values.targetMigration }}` when the value is non-empty
  - `restartPolicy: Never`, `backoffLimit: 0`
  - `ttlSecondsAfterFinished` set to a sensible default (e.g. 300)

**Acceptance Criteria**
- [ ] `helm template` with different `releaseIndex` values produces different Job names
- [ ] `helm template` with `targetMigration` set includes `--target` arg in the container command
- [ ] `helm template` with `targetMigration` omitted or empty does not include `--target` arg
- [ ] Job references the correct Secret for database connection environment variables
- [ ] Chart is reusable across apps (no app-specific hardcoding)

**Dependencies:** Task 8 (the base image must exist for the Job to reference)

---

## 10: DB Migrations — Reusable Tilt functions

Two Tilt utility functions: one for building the base migration Docker image, and one for building an app-specific migration image and deploying the migration Job alongside the application.

**Scope**
- `tools/tilt/utils.tiltfile`
  - `migrations_base_image()` — builds the base migration Docker image using `tools/docker/migrations.Dockerfile` with `apps/platform/db-migrations/` as context
  - `migrations_deploy(name, namespace, migrations_dir, db_secret_name, release_index, target_migration, deps)` — builds an app-specific Docker image (FROM base, COPY app's SQL files into `/migrations`), deploys the common Helm chart with the app's values via `k8s_yaml(helm(...))`, registers the Job as a `k8s_resource` with `resource_deps`
- `Tiltfile` (root)
  - Load and call `migrations_base_image()` so the base image is available to all apps
- `apps/platform/db-migrations/Tiltfile`
  - Standalone entrypoint that exercises both functions for testing the framework itself (deploys a `PostgresDatabase` + `PostgresCredential` via db-operator, runs sample migrations)

**Acceptance Criteria**
- [ ] `migrations_base_image()` builds the base Docker image successfully
- [ ] `migrations_deploy()` builds an app-specific image extending the base and deploys the Helm chart Job
- [ ] Both functions are loadable from any app Tiltfile via standard `load()` syntax
- [ ] App-specific migration Job deployment integrates with `resource_deps` so it runs after its dependencies are ready
- [ ] Root Tiltfile calls `migrations_base_image()` so the base image is available cluster-wide

**Dependencies:** Task 8 (Go runner and base Docker image) and Task 9 (common Helm chart)
