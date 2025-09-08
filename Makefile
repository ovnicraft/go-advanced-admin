SHELL := /bin/bash
GO := go

.PHONY: fmt format vet tests test

# Format all Go code in the repository
fmt:
	@echo "Formatting Go code (gofmt + go fmt)"
	@gofmt -s -w .
	@$(GO) fmt ./...

# Alias for fmt
format: fmt
	@:

vet:
	@echo "Vet code..."
	@$(GO) vet ./...

# Run tests via Docker Compose (parity with CI)
tests:
	@echo "Running tests with Docker Compose"
	@DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker compose up --build --abort-on-container-exit --exit-code-from root-tests; \
	status=$$?; \
	docker compose down -v --remove-orphans || true; \
	exit $$status

# Alias
test: tests
	@:
