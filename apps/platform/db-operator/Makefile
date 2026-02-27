# Image URL to use for all building/pushing image targets
IMG ?= db-operator:latest

# Root of the repository (used to locate shared Dockerfiles)
REPO_ROOT ?= $(shell git rev-parse --show-toplevel 2>/dev/null || echo $(abspath $(dir $(firstword $(MAKEFILE_LIST)))/../../..))

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be used in targets.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: ## Generate code (DeepCopy methods, CRD manifests, etc.)
	go generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -v -coverprofile cover.out

.PHONY: lint
lint: ## Run golangci-lint against code.
	@if command -v golangci-lint &>/dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed — running go vet as fallback"; \
		go vet ./...; \
	fi

##@ Build

.PHONY: build
build: fmt vet ## Build the operator binary.
	go build -o bin/db-operator ./cmd/

.PHONY: run
run: fmt vet ## Run the operator locally against the configured Kubernetes cluster.
	go run ./cmd/

.PHONY: image
image: ## Build and tag the db-operator container image.
	docker build \
		-f $(REPO_ROOT)/tools/docker/golang.Dockerfile \
		--build-arg BINARY=db-operator \
		--build-arg CMD_PATH=./cmd \
		-t $(IMG) \
		.

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf bin/ cover.out
