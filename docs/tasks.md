# Tasks

An itemised list of the next most important changes to make, ordered by priority.

| # | Title | Priority | Status |
|---|-------|----------|--------|
| 4 | [DB Operator — PostgresDatabaseReconciler](#4-db-operator--postgresdatabasereconciler) | High | ✅ Done |
| 5 | [DB Operator — PostgresDatabase Admin Secret](#5-db-operator--postgresdatabase-admin-secret) | High | ✅ Done |
| 7 | [DB Operator — Avoid redundant StatefulSet re-read after write](#7-db-operator--avoid-redundant-statefulset-re-read-after-write) | High | Not Started |
| 8 | [DB Operator — Add drift check before StatefulSet update](#8-db-operator--add-drift-check-before-statefulset-update) | High | Not Started |
| 9 | [DB Operator — Batch status sub-resource updates](#9-db-operator--batch-status-sub-resource-updates) | Medium | Not Started |
| 6 | [DB Operator — PostgresCredentialReconciler](#6-db-operator--postgrescredentialreconciler) | High | Not Started |

---

## ✅ 4: DB Operator — PostgresDatabaseReconciler

Implement the `PostgresDatabaseReconciler` in `apps/platform/db-operator/internal/controller/`. This controller is solely responsible for owning and reconciling the Kubernetes workloads (a `StatefulSet` and a headless `Service`) that back a `PostgresDatabase` CR. Validate with integration tests against a live Postgres instance.

**Scope**
- `PostgresDatabaseReconciler` in `apps/platform/db-operator/internal/controller/`
  - Creates and owns a `StatefulSet` (postgres container, PVC) and headless `Service`
  - Updates `PostgresDatabaseStatus` phase through `Pending` → `Ready` / `Failed`
  - Finalizer to cascade-delete owned `StatefulSet` and `Service` on CR deletion
  - Registered in `cmd/main.go`
- Integration tests covering the full lifecycle: create → ready → delete

**Acceptance Criteria**
- [x] Applying a `PostgresDatabase` CR results in a healthy Postgres `StatefulSet` and `Service`
- [x] `PostgresDatabaseStatus.phase` transitions correctly through `Pending` → `Ready`
- [x] Deleting a `PostgresDatabase` CR cascades deletion to its `StatefulSet` and `Service`
- [x] No orphaned Kubernetes resources remain after deletion
- [x] Integration tests cover all above lifecycle transitions and pass

**Dependencies:** Task 3 (CRD types) must be completed first

---

## ✅ 5: DB Operator — PostgresDatabase Admin Secret

Update the `PostgresDatabaseReconciler` to generate a random admin password for each `PostgresDatabase` and store the superuser credentials in an owned Kubernetes `Secret`. Replace the current hardcoded `POSTGRES_PASSWORD` with a `secretKeyRef` pointing at the generated Secret. This provides secure admin access for debugging and a reliable credential source for the `PostgresCredentialReconciler` (task 6) to connect to Postgres when provisioning application users.

**Scope**
- `PostgresDatabaseStatus` in `internal/api/v1alpha1/postgresdatabase_types.go`
  - Add a `SecretName string` status field to record the name of the admin credentials Secret
  - Regenerate CRD manifests via `make generate`
- `PostgresDatabaseReconciler` in `internal/controller/postgresdatabase_controller.go`
  - New `reconcileAdminSecret` step that runs before `reconcileStatefulSet`
    - On first reconcile: generate a random password (`crypto/rand`, 24+ characters, alphanumeric), create a `Secret` with keys `username` (value `postgres`) and `password`, set controller owner reference
    - On subsequent reconciles: verify the Secret still exists; recreate if missing but **do not** rotate the password if the Secret is present (password stability is required so the running Postgres instance stays accessible)
  - Update `desiredStatefulSet` to source `POSTGRES_PASSWORD` from the admin Secret via `secretKeyRef` instead of a literal value
  - Update `reconcileDelete` to delete the admin Secret alongside the StatefulSet and Service
  - Add `Owns(&corev1.Secret{})` in `SetupWithManager` so Secret changes trigger reconciliation
  - Add RBAC markers for `""` group `secrets` resource (`get;list;watch;create;update;patch;delete`)
  - Populate `PostgresDatabaseStatus.SecretName` after the Secret is created
- Integration tests in `internal/controller/postgresdatabase_controller_test.go`
  - Verify the admin Secret is created with `username` and `password` keys when a `PostgresDatabase` CR is applied
  - Verify the Secret has a controller owner reference pointing at the `PostgresDatabase` CR
  - Verify `PostgresDatabaseStatus.SecretName` is populated
  - Verify the StatefulSet container sources `POSTGRES_PASSWORD` from the Secret (not a literal)
  - Verify the Secret is deleted when the `PostgresDatabase` CR is deleted

**Acceptance Criteria**
- [x] A `Secret` containing `username` and `password` keys is created alongside each `PostgresDatabase`
- [x] The Secret has a controller owner reference to the `PostgresDatabase` CR
- [x] `PostgresDatabaseStatus.SecretName` is populated with the admin Secret name
- [x] The StatefulSet's `POSTGRES_PASSWORD` env var is sourced via `secretKeyRef` (no hardcoded password)
- [x] The admin Secret is deleted during CR finalizer cleanup
- [x] The database still reaches `Ready` phase with the Secret-backed password
- [x] All existing and new integration tests pass

**Dependencies:** Task 4 (initial `PostgresDatabaseReconciler`) must be completed first

---

## 7: DB Operator — Avoid redundant StatefulSet re-read after write

After `reconcileStatefulSet` creates or updates the StatefulSet, `updatePhaseFromStatefulSet` immediately performs a redundant `r.Get()` to fetch the same object. Per the Kubernetes standards, write responses already fill the in-memory object with the latest API server state; re-reading from the cache may return stale data.

Refactor `reconcileStatefulSet` to return the `*appsv1.StatefulSet` it created or updated, and pass that directly to `updatePhaseFromStatefulSet` so it operates on the fresh write response rather than a cache read.

**Scope**
- `reconcileStatefulSet` in `internal/controller/postgresdatabase_controller.go`
  - Change return signature to `(*appsv1.StatefulSet, error)`
  - Return the `desired` object after `r.Create`, or the `existing` object after `r.Update`
- `updatePhaseFromStatefulSet` in the same file
  - Change signature to accept a `*appsv1.StatefulSet` parameter instead of performing `r.Get()`
- `Reconcile` method
  - Thread the returned StatefulSet from `reconcileStatefulSet` into `updatePhaseFromStatefulSet`
- Existing integration tests must continue to pass

**Acceptance Criteria**
- [ ] `reconcileStatefulSet` returns the `*appsv1.StatefulSet` from the write response
- [ ] `updatePhaseFromStatefulSet` uses the passed-in StatefulSet — no `r.Get()` call for the StatefulSet
- [ ] All existing integration tests pass without modification

**Dependencies:** None (refactor of existing code)

**Standard:** `docs/standards/kubernetes.md` — "Don't read an object again if you just sent a write request"

---

## 8: DB Operator — Add drift check before StatefulSet update

`reconcileStatefulSet` unconditionally calls `r.Update()` on the StatefulSet every reconcile, even when the spec has not changed. This generates unnecessary etcd writes and watch events. The Service reconciler already has such a check; the StatefulSet reconciler should follow the same pattern.

Add an `equality.Semantic.DeepEqual` check on the StatefulSet's `Spec.Template` before calling `r.Update`, matching the pattern already used in `reconcileService`.

**Scope**
- `reconcileStatefulSet` in `internal/controller/postgresdatabase_controller.go`
  - After fetching the existing StatefulSet, compare `existing.Spec.Template` with `desired.Spec.Template` using `equality.Semantic.DeepEqual`
  - Only call `r.Update` when they differ
  - When they are equal, return the existing object without issuing a write
- Existing integration tests must continue to pass

**Acceptance Criteria**
- [ ] `r.Update` is not called when the StatefulSet spec template has not drifted
- [ ] `r.Update` is still called when the spec template has changed
- [ ] All existing integration tests pass without modification

**Dependencies:** Task 7 is recommended first (return signature change) but not strictly required

**Standard:** `docs/standards/kubernetes.md` — "Leverage optimistic locking" / minimise unnecessary writes

---

## 9: DB Operator — Batch status sub-resource updates

A single reconcile cycle can issue two separate `r.Status().Update()` calls: one in `reconcileAdminSecret` (setting `status.secretName`) and one in `setPhase` (setting `status.phase` and `status.conditions`). Each status update is a quorum write to etcd. Consolidate these into a single status write at the end of the reconcile loop.

**Scope**
- `reconcileAdminSecret` in `internal/controller/postgresdatabase_controller.go`
  - Remove the inline `r.Status().Update()` call; mutate `pgdb.Status.SecretName` in memory only
- `Reconcile` method
  - Perform a single `r.Status().Update()` after all sub-reconcilers have run and `setPhase` has set its fields
- `setPhase`
  - Remove its own `r.Status().Update()` call; only mutate the in-memory status fields
  - The caller (`Reconcile`) is responsible for persisting status
- Existing integration tests must continue to pass

**Acceptance Criteria**
- [ ] Only one `r.Status().Update()` call occurs per reconcile cycle (in the main `Reconcile` method)
- [ ] `status.secretName`, `status.phase`, and `status.conditions` are all correctly persisted
- [ ] All existing integration tests pass without modification

**Dependencies:** Tasks 7 and 8 should be completed first

**Standard:** `docs/standards/kubernetes.md` — "every direct API call results in a quorum read from etcd, which can be costly"

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
