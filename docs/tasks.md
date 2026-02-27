# Tasks

An itemised list of the next most important changes to make, ordered by priority.

| # | Title | Priority | Status |
|---|-------|----------|--------|
| 1 | [Root Makefile — Local Kubernetes Cluster (k3d)](#1-root-makefile--local-kubernetes-cluster-k3d) | High | ✅ Done |
| 2 | [DB Operator — Helm Chart & RBAC Manifests](#2-db-operator--helm-chart--rbac-manifests) | High | ✅ Done |
| 3 | [DB Operator — PostgresDatabase CRD API Types](#3-db-operator--postgresdatabase-crd-api-types) | High | Not Started |
| 4 | [DB Operator — PostgresDatabase Controller & Integration Tests](#4-db-operator--postgresdatabase-controller--integration-tests) | High | Not Started |

---

## ✅ 1: Root Makefile — Local Kubernetes Cluster (k3d)

Create the root `Makefile` and a `k3d` cluster configuration YAML at the repository root. The Makefile must expose `cluster-up` and `cluster-down` targets so any developer or CI agent can spin up and tear down a local Kubernetes cluster in a single command. The cluster should also include a local image registry so operator images can be pushed and pulled without a remote registry.

**Scope**
- `k3d-config.yaml` at repo root defining the cluster (name, node count, registry, port mappings)
- Root `Makefile` with `cluster-up`, `cluster-down`, and `help` targets
- Registry wired into the k3d cluster so `localhost:5000/<image>` resolves inside pods
- kubeconfig created in a `~/.scratch` directory as part of `cluster-up`, set as the target config with direnv to avoid accidentally applying things to the wrong cluster

**Acceptance Criteria**
- [x] `make cluster-up` creates a healthy k3d cluster and local registry
- [x] `make cluster-down` tears the cluster and registry down cleanly
- [x] `kubectl get nodes` shows all nodes `Ready` after `cluster-up`
- [x] Images pushed to the local registry are pullable from within the cluster
- [x] Cluster name and config path are overridable via Makefile variables
- [x] `make help` lists all targets with descriptions

**Dependencies:** None

---

## ✅ 2: DB Operator — Helm Chart & RBAC Manifests

Author the Helm chart that deploys the `db-operator` into a Kubernetes cluster, including all RBAC resources required for the operator to reconcile `PostgresDatabase` and `PostgresCredential` custom resources and manage the Kubernetes workloads they own.

**Scope**
- `apps/platform/db-operator/helm/` chart directory with `Chart.yaml`, `values.yaml`, and templates
- `Deployment` template for the operator pod
- `ServiceAccount`, `ClusterRole`, and `ClusterRoleBinding` templates
- CRD install templates (or a `crds/` directory)
- `Tiltfile` at the repo root (or `apps/platform/db-operator/Tiltfile`) wiring up image build and `helm upgrade --install` for local development

The `ClusterRole` must grant the operator least-privilege access covering:
- `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` on `postgresdatabases` and `postgrescredentials` resources and their `status` subresources
- `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` on `statefulsets` and `services` in `apps`
- `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` on `secrets` (for credential management)
- `get`, `list`, `watch` on `events`

**Acceptance Criteria**
- [x] `helm template` renders all manifests without errors
- [ ] `helm install` deploys the operator successfully to a local k3d cluster
- [ ] Operator pod reaches `Running` state and liveness/readiness probes pass
- [x] `ClusterRole` is least-privilege and covers all resources the controller reconciles
- [x] Image repository and tag are fully configurable via `values.yaml`
- [ ] `tilt up` builds and deploys the operator to the local cluster end-to-end

**Dependencies:** Task 1 (k3d cluster) must be completed first

---

## 3: DB Operator — PostgresDatabase CRD API Types

Define the Go API types, kubebuilder validation markers, status subresource, and generated CRD manifests for the two Postgres custom resources: `PostgresDatabase` (represents a Postgres instance) and `PostgresCredential` (represents a database user and its Kubernetes `Secret`). These form the schema contract the controller in task 4 reconciles against.

**Scope**
- API types in `apps/platform/db-operator/internal/api/v1alpha1/` (or equivalent versioned package)
- `PostgresDatabaseSpec`: database name, Postgres version, storage size
- `PostgresDatabaseStatus`: phase (`Pending` | `Ready` | `Failed`), conditions
- `PostgresCredentialSpec`: target `PostgresDatabase` reference, username, secret name, database-level permissions
- `PostgresCredentialStatus`: phase, conditions, secret name
- `+kubebuilder` markers for validation, status subresource, and printer columns
- `make generate` produces deep-copy methods and CRD YAML manifests committed to the chart's `crds/` directory

**Acceptance Criteria**
- [ ] `kubectl apply -f` on the generated CRD manifests succeeds without errors
- [ ] `kubectl get postgresdatabases` and `kubectl get postgrescredentials` return the correct group/version
- [ ] All spec fields have validation markers (required/optional, min/max, enum where applicable)
- [ ] Status subresource is enabled on both CRDs
- [ ] Generated deep-copy code and CRD manifests are committed
- [ ] CRD manifests are present in the Helm chart under `crds/`

**Dependencies:** None

---

## 4: DB Operator — PostgresDatabase Controller & Integration Tests

Implement the reconciliation controllers for both `PostgresDatabase` and `PostgresCredential` CRDs. The `PostgresDatabase` controller owns a `StatefulSet` and headless `Service`. The `PostgresCredential` controller creates a Postgres user inside the target database and writes the credentials into a Kubernetes `Secret`. Validate both controllers with integration tests against a real PostgreSQL instance, per the project testing standards.

**Scope**
- `PostgresDatabaseReconciler` in `apps/platform/db-operator/internal/controller/`
  - Creates and owns a `StatefulSet` (postgres container, PVC) and headless `Service`
  - Updates `PostgresDatabaseStatus` phase through `Pending` → `Ready` / `Failed`
  - Finalizer to cascade-delete owned resources on CR deletion
- `PostgresCredentialReconciler` in `apps/platform/db-operator/internal/controller/`
  - Waits for the target `PostgresDatabase` to be `Ready`
  - Connects to Postgres and creates the specified user with a randomised password
  - Writes username and password into the named Kubernetes `Secret`
  - Updates `PostgresCredentialStatus` phase and secret reference
  - Finalizer to drop the Postgres user and delete the `Secret` on CR deletion
- Integration tests in `apps/platform/db-operator/internal/controller/` using `envtest` or a live Postgres instance spun up via Docker/Tilt

**Acceptance Criteria**
- [ ] Applying a `PostgresDatabase` CR results in a healthy Postgres `StatefulSet` and `Service`
- [ ] `PostgresDatabaseStatus.phase` transitions correctly through `Pending` → `Ready`
- [ ] Applying a `PostgresCredential` CR (with a `Ready` database) creates the Postgres user and populates the named `Secret`
- [ ] `PostgresCredentialStatus.phase` transitions correctly to `Ready`
- [ ] Deleting a `PostgresDatabase` CR cascades deletion to its `StatefulSet` and `Service`
- [ ] Deleting a `PostgresCredential` CR drops the Postgres user and removes the `Secret`
- [ ] Integration tests cover all above lifecycle transitions and pass in CI
- [ ] No orphaned Kubernetes resources or Postgres users remain after deletion

**Dependencies:** Task 3 (CRD types) must be completed first; task 2 (Helm chart) should be complete for in-cluster testing
