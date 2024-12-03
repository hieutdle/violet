# Makefile

.PHONY: test
test:
	@./scripts/unit-tests.sh

.PHONY: format 
format:
	@./scripts/format.sh


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
	@echo "       make build"
	@echo "       make install"
	