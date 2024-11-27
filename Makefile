# Makefile

.PHONY: test

test:
	@echo "Running tests..."
	@go test ./...
	@echo "All tests passed."

.PHONY: format 

format:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted."

.PHONY: build

build:
	@./scripts/build.sh

.PHONY: install

install:
	@./scripts/install.sh

.PHONY: usage 

usage:
	@echo "Usage: make test"
	@echo "       make format"

