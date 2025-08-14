SHELL := /usr/bin/env bash
E2E_TAG ?= e2e

REGISTRY_NAME ?= registry

GOLANGCI_LINT=golangci-lint
TIMEOUT_UNIT = 20m
GOFUMPT=gofumpt

BIN = $(CURDIR)/.bin
# release directory where the Tekton resources are rendered into.
RELEASE_VERSION=v0.1.1
RELEASE_DIR ?= /tmp/tekton-caches-${RELEASE_VERSION}
$(BIN):
	@mkdir -p $@
CATALOGCD = $(or ${CATALOGCD_BIN},${CATALOGCD_BIN},$(BIN)/catalog-cd)
$(BIN)/catalog-cd: $(BIN)
	curl -fsL https://github.com/openshift-pipelines/catalog-cd/releases/download/v0.3.0/catalog-cd_0.3.0_linux_x86_64.tar.gz | tar xzf - -C $(BIN) catalog-cd



e2e-coverage: ## run e2e tests with coverage
	tests/e2e.sh
	@go test -v -failfast -count=1 -tags=$(E2E_TAG) ./tests/ -coverpkg=./... -coverprofile /tmp/coverage.out
	@go tool cover -func /tmp/coverage.out

e2e: e2e-coverage

e2e-docker: ## run e2e tests with a docker registry started
	@if [[ $$(docker ps --filter name=$(REGISTRY_NAME) --format '{{.ID}}') == "" ]]; then \
		echo -n "Starting container $(REGISTRY_NAME) ... "; \
		docker run  --name $(REGISTRY_NAME) -p 127.0.0.1:5000:5000 -d registry:2 >/dev/null; \
		echo "done"; \
	fi
	make e2e-coverage
	@docker rm -f $(REGISTRY_NAME) >/dev/null

lint: lint-go
lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
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
	gh release create $(RELEASE_VERSION) --generate-notes && \
	gh release upload $(RELEASE_VERSION) $(RELEASE_DIR)/release/catalog.yaml && \
	gh release upload $(RELEASE_VERSION) $(RELEASE_DIR)/release/resources.tar.gz