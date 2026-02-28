# Tiltfile — games-hub root entrypoint
#
# Loads deployment functions from each application and calls them.
# Integration tests are run by each app's own Tiltfile entrypoint.
# End-to-end tests will be added here as additional services come online.
#
# Prerequisites:
#   - k3d cluster running (make cluster-up from repo root)
#   - KUBECONFIG pointing at ~/.scratch/games-hub.yaml
#
# Usage:
#   tilt up

# ── Application deployments ───────────────────────────────────────────────────

load("apps/platform/db-operator/Tiltfile", "deploy_db_operator")

deploy_db_operator()

# ── End-to-end tests (placeholder) ───────────────────────────────────────────
# local_resource("e2e-tests", cmd="make e2e-test", resource_deps=[...])
