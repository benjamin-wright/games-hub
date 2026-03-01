# Kubernetes Standards

## Controllers

**Governing principle:** Assume all data is eventually consistent and may be stale — even data read directly from the API server.

- Use the informer cache for all reads. Fall back to a direct read only when caching is infeasible (high-churn objects, memory pressure); in that case, filter by namespace/labels and prefer metadata-only calls.
- Guard status writes with a state check — skip the write if nothing has changed.
- Use deterministic names for child objects so that optimistic locking detects conflicts naturally.
- Send updates/patches with the last-known `resourceVersion`. On conflict, return an error and let the work queue retry with backoff.
- After a write (Create/Update/Patch/Delete), use the object returned by the API server — do not re-read from the cache, which may contain an older `resourceVersion`.
- If using `generateName`, track outstanding creates in memory and retry if the expected watch event does not arrive within a reasonable timeout.
- Reuse existing caches (e.g. controller watches) rather than adding direct reads for the same object type.
