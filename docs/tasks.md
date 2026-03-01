# Tasks

An itemised list of the next most important changes to make, ordered by priority.

| # | Title | Priority | Status |
|---|-------|----------|--------|
| 9 | [DB Migrations — Common Helm chart](#9-db-migrations--common-helm-chart) | High | Todo |
| 10 | [DB Migrations — Reusable Tilt functions](#10-db-migrations--reusable-tilt-functions) | High | Todo |
| 11 | [DB Operator — Isolate standalone Tilt deployment from platform deployment](#11-db-operator--isolate-standalone-tilt-deployment-from-platform-deployment) | High | Todo |

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

---

## 11: DB Operator — Isolate standalone Tilt deployment from platform deployment

The db-operator's standalone local deployment (for integration testing) currently shares the same resources as the platform-wide deployment managed by the root Tiltfile. When run standalone, it replaces the platform deployment rather than coexisting alongside it. The standalone deployment must run as a separate, namespaced instance so both can operate simultaneously without conflict.

**Scope**
- `apps/platform/db-operator/helm/values.yaml`
  - Change `args.instanceName` default from `"default"` to `""` (empty); empty means the operator processes unlabeled CRs, which is the correct platform behaviour
- `apps/platform/db-operator/helm/templates/clusterrole.yaml` and `clusterrolebinding.yaml`
  - Include the release namespace as a suffix in cluster-scoped resource names so that parallel deployments in different namespaces do not collide
- `apps/platform/db-operator/Tiltfile`
  - `deploy_db_operator(namespace, instance_name)` — accepts a target namespace and an instance name; passes both through to the Helm release (`namespace`, `args.instanceName`); remove `port_forwards` from the `k8s_resource` call
  - Standalone block (`config.main_dir == _APP_DIR`) — calls `deploy_db_operator("db-operator-test", "test")`; `integration-tests` resource depends on the release name
  - Platform callers invoke `deploy_db_operator("db-operator", "")` for the default cluster-wide operator

**Acceptance Criteria**
- [ ] Running `tilt up` from the db-operator directory deploys to namespace `db-operator-test` with `--instance-name=test`, without touching the platform deployment in `db-operator`
- [ ] The root Tiltfile's `deploy_db_operator("db-operator", "")` deploys the platform-wide operator with an empty instance name
- [ ] Both deployments can run simultaneously — different namespaces, namespace-suffixed cluster-scoped resources, no port-forward conflicts
- [ ] Integration tests create CRs with label `games-hub.io/operator-instance: test` and are reconciled only by the test instance

**Dependencies:** None
