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

.PHONY: format-scripts
format-scripts:
	@find scripts/ -type f -exec sed -i 's/\r$$//' {} +

.PHONY: usage 
usage:
	@echo "Usage: make test"
	@echo "       make format"
	@echo "       make build"
	@echo "       make install"
	@echo "       make format-scripts"
