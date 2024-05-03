E2E_TAG ?= e2e
REGISTRY_NAME ?= registry

e2e-coverage:
	@go test -failfast -count=1 -tags=$(E2E_TAG) ./tests -coverpkg=./... -coverprofile /tmp/coverage.out
	@go tool cover -func /tmp/coverage.out
e2e: e2e-coverage

e2e-docker:
	@if [[ $$(docker ps --filter name=$(REGISTRY_NAME) --format '{{.ID}}') == "" ]]; then \
		echo -n "Starting container $(REGISTRY_NAME) ... "; \
		docker run  --name $(REGISTRY_NAME) -p 127.0.0.1:5000:5000 -d registry:2 >/dev/null; \
		echo "done"; \
	fi
	make e2e-coverage
	@docker rm -f $(REGISTRY_NAME) >/dev/null
