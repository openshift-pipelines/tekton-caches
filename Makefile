SHELL := /usr/bin/env bash
E2E_TAG ?= e2e

REGISTRY_NAME ?= registry

TIMEOUT_UNIT = 20m
GOFUMPT=gofumpt
TEKTON_VERSION ?= $(shell go list -m -f "{{ .Version }}" github.com/tektoncd/pipeline)
KUBECTL ?= kubectl

BIN = $(CURDIR)/.bin
# release directory where the Tekton resources are rendered into.
RELEASE_VERSION=v0.1.1
RELEASE_DIR ?= /tmp/tekton-caches-${RELEASE_VERSION}
$(BIN):
	@mkdir -p $@
CATALOGCD = $(or ${CATALOGCD_BIN},${CATALOGCD_BIN},$(BIN)/catalog-cd)
$(BIN)/catalog-cd: $(BIN)
	curl -fsL https://github.com/openshift-pipelines/catalog-cd/releases/download/v0.3.0/catalog-cd_0.3.0_linux_x86_64.tar.gz | tar xzf - -C $(BIN) catalog-cd

# The go version this module targets, read straight from go.mod.
GO_MOD_VERSION = $(shell go list -m -f '{{.GoVersion}}')

# golangci-lint refuses to load its config when it was built with a go
# toolchain older than the one go.mod targets, and an older release can also
# fail to decode export data produced by a newer toolchain. A published binary
# is built with whichever go version that release used, so we cannot pick one
# by go version -- instead we build golangci-lint from source with this repo's
# own toolchain, which makes it track GO_MOD_VERSION automatically.
#
# Override GOLANGCI_LINT to use an installation you manage yourself.
GOLANGCI_LINT_VERSION ?= v2.15.0
ifndef GOLANGCI_LINT
GOLANGCI_LINT = $(BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
# Only install when we own the binary; an override is the caller's to provide.
GOLANGCI_LINT_INSTALL = $(GOLANGCI_LINT)
endif

# GOFLAGS is cleared because this repo vendors its dependencies, and
# "go install pkg@version" refuses to run under -mod=vendor.
$(BIN)/golangci-lint-%: go.mod | $(BIN)
	@echo "Building golangci-lint $* with $$(go env GOVERSION) (go.mod targets go$(GO_MOD_VERSION))..."
	@GOFLAGS= GOBIN=$(BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$*
	@mv $(BIN)/golangci-lint $@



e2e-coverage:  ## run e2e tests with coverage
	tests/e2e.sh
	@go test -v -failfast -count=1 -tags=$(E2E_TAG) ./tests/ -coverpkg=./... -coverprofile /tmp/coverage.out
	@go tool cover -func /tmp/coverage.out

e2e: kind tekton e2e-coverage

e2e-docker: ## run e2e tests with a docker registry started
	@if [[ $$(docker ps --filter name=$(REGISTRY_NAME) --format '{{.ID}}') == "" ]]; then \
		echo -n "Starting container $(REGISTRY_NAME) ... "; \
		docker run  --name $(REGISTRY_NAME) -p 127.0.0.1:5000:5000 -d registry:2 >/dev/null; \
		echo "done"; \
	fi
	make e2e-coverage
	@docker rm -f $(REGISTRY_NAME) >/dev/null

PKG ?= ./...
lint: lint-go
lint-go: $(GOLANGCI_LINT_INSTALL) ## runs go linter on all go files
	@built=$$($(GOLANGCI_LINT) version 2>/dev/null | sed -n 's/.*built with go\([0-9.]*\).*/\1/p'); \
	if [ -n "$$built" ] && [ "$$(printf '%s\n%s\n' "$(GO_MOD_VERSION)" "$$built" | sort -V | tail -1)" != "$$built" ]; then \
		echo "$(GOLANGCI_LINT) was built with go$$built, but go.mod targets go$(GO_MOD_VERSION)."; \
		echo "It cannot load the config. Unset GOLANGCI_LINT to build a compatible one."; \
		exit 1; \
	fi
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run $(PKG) --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--timeout $(TIMEOUT_UNIT)

unit: unit-tests
unit-tests: ## runs unit tests
	@echo "Running Unit tests..."
	go test -race ./...

.PHONY: fumpt ## formats the GO code with gofumpt(excludes vendors dir)
fumpt:
	@find internal cmd tests -name '*.go'|xargs -P4 $(GOFUMPT) -w -extra

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

#Release
# pepare a release
.PHONY: prepare-release
prepare-release:
	mkdir -p $(RELEASE_DIR) || true
	hack/release.sh $(RELEASE_DIR)


.PHONY: release
release: ${CATALOGCD} prepare-release
	pushd ${RELEASE_DIR} && \
		$(CATALOGCD) release \
			--output release \
			--version $(RELEASE_VERSION) \
			stepactions/* \
		; \
	popd

# tags the repository with the RELEASE_VERSION and pushes to "origin"
git-tag-release-version:
	if ! git rev-list "${RELEASE_VERSION}".. >/dev/null; then \
		git tag "$(RELEASE_VERSION)" && \
			git push origin --tags; \
	fi

# github-release
.PHONY: github-release
github-release: git-tag-release-version release
	if ! gh release view $(RELEASE_VERSION) >/dev/null 2>&1; then \
		gh release create $(RELEASE_VERSION) --generate-notes; \
	fi && \
	gh release upload $(RELEASE_VERSION) $(RELEASE_DIR)/release/catalog.yaml --clobber && \
	gh release upload $(RELEASE_VERSION) $(RELEASE_DIR)/release/resources.tar.gz --clobber


.PHONY: gitlint
gitlint: ## Lint commit messages (last commit; set GITLINT_COMMITS=origin/main..HEAD for a range)
	@command -v gitlint >/dev/null 2>&1 || { \
		echo "gitlint not found. Install with: pip install gitlint"; \
		exit 1; \
	}
	@if [ -n "$(GITLINT_COMMITS)" ]; then \
		gitlint --commits "$(GITLINT_COMMITS)"; \
	else \
		gitlint; \
	fi

.PHNY: go-build
go-build:
	go build -tags strictfipsruntime  -v -o /tmp/cache  ./cmd/cache

.PHONY: tekton
tekton:
	$(KUBECTL) apply --server-side -f https://github.com/tektoncd/pipeline/releases/download/$(TEKTON_VERSION)/release.yaml
	$(KUBECTL) wait --for=condition=Available deployment --all -n tekton-pipelines --timeout=300s

.PHONY: kind
kind:
	set -x
	./hack/kind-install.sh
	kind export kubeconfig
