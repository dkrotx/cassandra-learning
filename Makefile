# Make the default target show help
.DEFAULT_GOAL := help

.PHONY: start-cassandra
.PHONY: stop-cassandra
.PHONY: install-schema
.PHONY: help

start-cassandra: ## Start the Cassandra container and install schema
	cd docker && docker-compose up -d

stop-cassandra: ## Stop the Cassandra container
	cd docker && docker-compose down

install-schema: ## install schema to running Cassandra
	docker run -v ./schema:/schema --entrypoint '/bin/sh' cassandra:latest -c "cqlsh host.docker.internal -f /schema/myjournal.cql && echo SUCCESS || echo FALED"

help: ## Show this help
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*## "}; \
	     /^[a-zA-Z0-9_.-]+:.*## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' \
	     $(MAKEFILE_LIST) | sort
