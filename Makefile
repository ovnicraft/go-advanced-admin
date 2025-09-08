SHELL := /bin/bash
GO := go

.PHONY: fmt format vet

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
