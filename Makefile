# Make the default target show help
.DEFAULT_GOAL := help

.PHONY: start-cassandra
.PHONY: stop-cassandra
.PHONY: install-schema
.PHONY: bins
.PHONY: clean
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

help: ## Show this help
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*## "}; \
	     /^[a-zA-Z0-9_.-]+:.*## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' \
	     $(MAKEFILE_LIST) | sort
