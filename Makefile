# Make the default target show help
.DEFAULT_GOAL := help

.PHONY: start-cassandra
.PHONY: stop-cassandra
.PHONY: install-schema
.PHONY: bins
.PHONY: clean
.PHONY: tidy
.PHONY: fmt
.PHONY: lint
.PHONY: test
.PHONY: pr
.PHONY: help

start-cassandra: ## Start the Cassandra container and install schema
	cd docker && docker-compose up -d

stop-cassandra: ## Stop the Cassandra container
	cd docker && docker-compose down

install-schema: ## Install schema to running Cassandra
	docker run -v ./schema:/schema --entrypoint '/bin/sh' cassandra:latest -c "cqlsh host.docker.internal -f /schema/myjournal.cql && echo SUCCESS || echo FALED"
	
bins: ## Build the application binaries
	go build -o myjournal-cli ./cmd/cli

clean: ## Remove the binary files
	rm -f myjournal-cli

tidy: ## go mod tidy
	@echo Running go mod tidy for all packages
	@go mod tidy

fmt: ## Format the source files
	@echo Formatting source files
	@golangci-lint fmt ./...

lint: ## Lint the source files
	@echo Linting source files
	@golangci-lint run ./...

test: ## Run the tests
	go test -v ./...

pr: ## Run tests necessary before submitting PR
	$(MAKE) tidy
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) bins
	@echo all checks succeeded, PR is ready to be made

help: ## Show this help
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*## "}; \
	     /^[a-zA-Z0-9_.-]+:.*## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' \
	     $(MAKEFILE_LIST) | sort
