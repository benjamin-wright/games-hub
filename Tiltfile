# Tiltfile — db-operator local development
#
# Prerequisites:
#   - k3d cluster running (make cluster-up from repo root)
#   - KUBECONFIG pointing at ~/.scratch/games-hub.yaml
#
# Usage:
#   tilt up

# ── Settings ──────────────────────────────────────────────────────────────────

REGISTRY     = "localhost:5001"
IMAGE_NAME   = "db-operator"
IMAGE_REF    = "{}/{}".format(REGISTRY, IMAGE_NAME)

CHART_DIR    = "./apps/platform/db-operator/helm"
RELEASE_NAME = "db-operator"
NAMESPACE    = "db-operator"

# -- Custom Tilt extensions (optional) ─────────────────────────────────────────────

def namespace_create(name):
    """Create a Kubernetes namespace if it doesn't already exist."""
    k8s_yaml(blob("""
apiVersion: v1
kind: Namespace
metadata:
  name: {}
""".format(name)))

# ── Image build ───────────────────────────────────────────────────────────────

docker_build(
    IMAGE_REF,
    context    = "./apps/platform/db-operator",
    dockerfile = "./tools/docker/golang.Dockerfile",
    build_args = {
        "CMD_PATH": "./cmd",
        "BINARY":   "db-operator",
    },
    # Only rebuild when these paths change
    only = [
        "./apps/platform/db-operator",
        "./tools/docker/golang.Dockerfile",
    ],
)

# ── Namespace ─────────────────────────────────────────────────────────────────

namespace_create(NAMESPACE)

# ── Helm deploy ───────────────────────────────────────────────────────────────

k8s_yaml(
    helm(
        CHART_DIR,
        name      = RELEASE_NAME,
        namespace = NAMESPACE,
        set = [
            "image.repository={}".format(IMAGE_REF),
            "image.tag=latest",
            "image.pullPolicy=Always",
        ],
    )
)

# ── Resource configuration ────────────────────────────────────────────────────

k8s_resource(
    RELEASE_NAME,
    port_forwards = [
        "8080:8080",  # metrics
        "8081:8081",  # health probes
    ],
    labels = ["db-operator"],
)
