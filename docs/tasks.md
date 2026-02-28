# Tasks

An itemised list of the next most important changes to make, ordered by priority.

| # | Title | Priority | Status |
|---|-------|----------|--------|
| 7 | [DB Operator — Instance-scoped label filtering](#7-db-operator--instance-scoped-label-filtering) | High | Not Started |
| 2 | [DB Operator — Refactor integration test suite harness](#2-db-operator--refactor-integration-test-suite-harness) | High | Not Started |
| 6 | [DB Operator — PostgresCredentialReconciler](#6-db-operator--postgrescredentialreconciler) | High | Not Started |

---

## 7: DB Operator — Instance-scoped label filtering

Allow multiple db-operator instances to coexist in the same Kubernetes cluster by scoping each operator to only reconcile CRs that carry a matching instance label (`games-hub.io/operator-instance`). Filtering is implemented at the cache layer via `cache.Options.ByObject` label selectors, pushing filtering to the API server's list/watch request so the operator never receives or caches CRs belonging to other instances.

**Scope**
- `apps/platform/db-operator/cmd/main.go`
  - Add `--instance-name` CLI flag (default: `"default"`)
  - Build a `labels.Selector` from `games-hub.io/operator-instance: <instanceName>` and configure `cache.Options.ByObject` for `PostgresDatabase` and `PostgresCredential` types
  - Pass `InstanceName` to the `PostgresDatabaseReconciler` struct
  - Update leader election ID to `fmt.Sprintf("db-operator-%s.games-hub.io", instanceName)` to avoid lock conflicts between instances
- `apps/platform/db-operator/internal/controller/postgresdatabase_controller.go`
  - Add `InstanceName string` field to the reconciler struct
  - Update `labelsForDatabase()` to include `"games-hub.io/operator-instance": instanceName` in the returned label set
  - Update all call sites (admin Secret, Service, StatefulSet builders) to pass the instance name through
- `apps/platform/db-operator/helm/values.yaml`
  - Add `args.instanceName: "default"`
- `apps/platform/db-operator/helm/templates/deployment.yaml`
  - Add `--instance-name={{ .Values.args.instanceName }}` to container args
- `apps/platform/db-operator/internal/controller/suite_test.go`
  - Set `InstanceName: "integration-test"` on the reconciler struct
  - Add `cache.Options.ByObject` with the matching label selector to the test manager
- `apps/platform/db-operator/internal/controller/postgresdatabase_controller_test.go`
  - Update `newTestResources` to add `"games-hub.io/operator-instance": "integration-test"` to every test CR
  - Add a test case that creates a CR **without** the instance label and verifies it is never reconciled
- `apps/platform/db-operator/spec.md`
  - Document the instance label requirement, `--instance-name` flag, and multi-instance deployment behaviour

**Acceptance Criteria**
- [ ] Operator accepts `--instance-name` flag with a default value of `"default"`
- [ ] Cache-level label selector is configured so the informer only receives CRs matching the operator's instance label
- [ ] Owned sub-resources (StatefulSet, Service, Secret) carry the `games-hub.io/operator-instance` label
- [ ] Leader election ID incorporates the instance name to prevent lock conflicts
- [ ] A CR without the instance label is not reconciled (verified by integration test)
- [ ] Helm chart exposes `args.instanceName` and passes it to the deployment
- [ ] All existing integration tests pass with the instance label applied to test CRs

**Dependencies:** None

---

## 2: DB Operator — Refactor integration test suite harness

Refactor `apps/platform/db-operator/internal/controller/suite_test.go` to remove the in-process `controller-runtime` manager. The test suite should connect to an already-deployed operator (deployed by Tilt) rather than spinning up its own controller. `BeforeSuite` retains CRD application and direct-client setup; it no longer registers or starts a reconciler in-process.

**Scope**
- `apps/platform/db-operator/internal/controller/suite_test.go`
  - Remove: `ctrl.NewManager`, `controller.PostgresDatabaseReconciler` registration, `go func() { mgr.Start }` goroutine, and all associated imports (`ctrl`, `metricsserver`, `controller` package)
  - Retain: kubeconfig resolution, `kubectl apply` CRD step, direct `k8sClient` creation
  - The `AfterSuite` can remain as-is (only calls `cancel()`)

**Acceptance Criteria**
- [ ] `suite_test.go` imports no `controller-runtime` manager or local `controller` package
- [ ] `BeforeSuite` does not start any in-process reconciler
- [ ] All existing integration tests in `postgresdatabase_controller_test.go` pass when the operator is deployed via Tilt (Task 1)
- [ ] `make integration-test` continues to work when invoked via the Tilt `local_resource`

**Dependencies:** Task 1 (per-application Tiltfile) must be in place so the operator is deployed before tests run

**Standard:** `docs/standards/testing.md` — "simplify test harnesses where possible by deploying the application into an integration testing namespace and testing against the deployed application"

---

## 6: DB Operator — PostgresCredentialReconciler

Implement the `PostgresCredentialReconciler` in `apps/platform/db-operator/internal/controller/`. This controller is solely responsible for provisioning a Postgres user inside a target `PostgresDatabase` instance and writing the generated credentials into a Kubernetes `Secret`. Validate with integration tests against a live Postgres instance.

**Scope**
- `PostgresCredentialReconciler` in `apps/platform/db-operator/internal/controller/`
  - Waits for the target `PostgresDatabase` to be `Ready` before acting
  - Reads the admin credentials from the `Secret` referenced by `PostgresDatabaseStatus.SecretName` to connect to Postgres
  - Connects to Postgres and creates the specified user with a randomised password
  - Writes username and password into the named Kubernetes `Secret`
  - Updates `PostgresCredentialStatus` phase and secret reference
  - Finalizer to drop the Postgres user and delete the `Secret` on CR deletion
  - Registered in `cmd/main.go`
- Integration tests covering the full lifecycle: create → ready → delete, including the dependency-wait behaviour when the target database is not yet `Ready`

**Acceptance Criteria**
- [ ] Applying a `PostgresCredential` CR (with a `Ready` database) creates the Postgres user and populates the named `Secret`
- [ ] `PostgresCredentialStatus.phase` transitions correctly to `Ready`
- [ ] Controller waits (requeues) when the target `PostgresDatabase` is not yet `Ready`
- [ ] Deleting a `PostgresCredential` CR drops the Postgres user and removes the `Secret`
- [ ] No orphaned Postgres users or Kubernetes `Secrets` remain after deletion
- [ ] Integration tests cover all above lifecycle transitions and pass

**Dependencies:** Task 5 (`PostgresDatabase Admin Secret`) must be completed first so the credential controller can securely obtain admin credentials
