coverage:
	@go test ./... -coverpkg=./... -coverprofile /tmp/coverage.out
	@go tool cover -func /tmp/coverage.out
test: coverage
