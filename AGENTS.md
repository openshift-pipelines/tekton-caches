# Tekton Caches
This is a tool to cache resources like go cache/maven or others on TektonCD
pipelines.

For more details see [README.md](./README.md ).

## Quick Start

```bash
  make build-go # Build the project
  make lint-go # One command to check for lint inssues
  make unit-tests # Run Unit Tests
  ./hack/kind-install.sh  # Setup Kind Environment for E2E testing
  make e2e   # Run tests to verify setup
```

## Build & Test Commands

```bash

#Build Commands
make go-build # Only build command you need to build the cache binary

# Test & Coverage
make e2e
make e2e-coverage ## run e2e tests with coverage
make e2e-docker: ## run e2e tests with a docker registry started

make unit-tests: ## runs unit tests
make fumpt ## formats the GO code with gofumpt(excludes vendors dir)

# Lint — must pass before every PR
make lint # golangci-lint + yamllint (all packages)
make lint-go # Go only, all packages
make lint-go PKG=./internal/provider/oci/...  # single package
make lint-go PKG=./internal/provider/oci/upload.go  # single file linting (fast)
make gitlint # optional: lint last commit message (pip install gitlint)
make gitlint GITLINT_COMMITS=origin/main..HEAD  # lint all commits on the branch

#Release
make release
```
