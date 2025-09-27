install-schema:
	docker run -v ./schema:/schema --entrypoint '/bin/sh' cassandra:latest -c "cqlsh host.docker.internal -f /schema/myjournal.cql && echo SUCCESS || echo FALED"
