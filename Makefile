E2E_TAG ?= e2e
REGISTRY_NAME ?= registry

GOLANGCI_LINT=golangci-lint
TIMEOUT_UNIT = 20m
GOFUMPT=gofumpt

e2e-coverage: ## run e2e tests with coverage
	@go test -failfast -count=1 -tags=$(E2E_TAG) ./tests -coverpkg=./... -coverprofile /tmp/coverage.out
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
	go test ./...

.PHONY: fumpt ## formats the GO code with gofumpt(excludes vendors dir)
fumpt:
	@find internal cmd tests -name '*.go'|xargs -P4 $(GOFUMPT) -w -extra

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

