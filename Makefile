# Cluster name — matches the name in k3d-config.yaml
CLUSTER_NAME     ?= games-hub

# Path to the k3d cluster config
K3D_CONFIG       ?= $(abspath $(dir $(firstword $(MAKEFILE_LIST)))/k3d-config.yaml)

# Directory and path for the kubeconfig written during cluster-up
KUBECONFIG_DIR   ?= $(HOME)/.scratch
KUBECONFIG_PATH  ?= $(KUBECONFIG_DIR)/$(CLUSTER_NAME).yaml

# Registry settings
REGISTRY_NAME    ?= $(CLUSTER_NAME)-registry.localhost
REGISTRY_PORT    ?= 5001

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Cluster

.PHONY: cluster-up
cluster-up: ## Create the local k3d cluster and registry, then write kubeconfig to KUBECONFIG_PATH.
	@echo "Creating kubeconfig directory $(KUBECONFIG_DIR) …"
	@mkdir -p "$(KUBECONFIG_DIR)"
	@echo "Creating k3d cluster '$(CLUSTER_NAME)' using config $(K3D_CONFIG) …"
	k3d cluster create $(CLUSTER_NAME) \
		--registry-create $(REGISTRY_NAME):0.0.0.0:$(REGISTRY_PORT) \
		--kubeconfig-update-default=false \
		-p "80:80@loadbalancer" \
		--wait;
	@echo "Writing kubeconfig to $(KUBECONFIG_PATH) …"
	k3d kubeconfig get "$(CLUSTER_NAME)" > "$(KUBECONFIG_PATH)"
	@echo ""
	@echo "Cluster is ready. Run the following (or use direnv) to target it:"
	@echo "  export KUBECONFIG=$(KUBECONFIG_PATH)"
	@echo ""
	@KUBECONFIG="$(KUBECONFIG_PATH)" kubectl get nodes

.PHONY: cluster-down
cluster-down: ## Tear down the local k3d cluster and registry.
	@echo "Deleting k3d cluster '$(CLUSTER_NAME)' …"
	k3d cluster delete "$(CLUSTER_NAME)"
	@echo "Cluster '$(CLUSTER_NAME)' deleted."
