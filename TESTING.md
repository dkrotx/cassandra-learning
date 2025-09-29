# Testing Guide for Cassandra Learning Project

## Testing Approach: Testcontainers (Single File)

**What it is**: Automated containerized testing with real Cassandra
**Requirements**: Docker, testcontainers-go library

```bash
# Install testcontainers (already done)
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/cassandra

# Run tests
go test ./pkg/myjournal/ -v
```

**Key Features**:
- **No port conflicts**: Uses random ports (not 9042)
- **Single container**: One Cassandra instance for all tests (faster)
- **Test isolation**: Data cleaned between tests (TRUNCATE)
- **Automatic cleanup**: Container destroyed after all tests
- **Modern API**: Uses `GenericContainer` (not deprecated `RunContainer`)
- **Reliable startup**: Uses Docker health check approach (`cqlsh -e 'describe keyspaces'`)
- **Proper timeouts**: 3-minute container timeout, 2-minute health check timeout
- **Schema from file**: Uses existing `schema/myjournal.cql` file (no duplication)
- **Test suite**: Uses `testify/suite` for efficient setup/teardown

**Pros**:
- Real database behavior
- Isolated test environment
- CI/CD friendly
- No manual setup required
- Reproducible
- Single test file to maintain
- No port conflicts with existing Cassandra
- **Much faster**: One container for all tests
- **Test isolation**: Data cleaned between tests

**Cons**:
- Requires Docker
- Slightly slower than mocks

## Running Tests

### Full Integration Tests (Default)
```bash
# Run all tests with testcontainers
go test ./pkg/myjournal/ -v
```

### Skip Integration Tests (Fast)
```bash
# Run only fast tests (skips testcontainers)
go test ./pkg/myjournal/ -short -v
```

### Run Specific Test
```bash
# Run only one test
go test ./pkg/myjournal/ -v -run TestCreatePost
```

## Test Structure

### Tests (`myjournal_test.go`)
- `TestCreatePost` - Tests post creation
- `TestReadPostsByUser` - Tests reading posts
- `TestDeletePost` - Tests deletion by post ID only
- `TestDeletePostByUser` - Tests deletion with user validation
- `TestAddTags` - Tests adding tags to posts
- `TestRemoveTags` - Tests removing tags from posts

All tests use a test suite with testcontainers for automated Cassandra setup and automatically load the schema from `schema/myjournal.cql`. The test suite creates one Cassandra container for all tests and cleans data between tests for isolation.

## Best Practices

### 1. Test Isolation
- Each test should be independent
- Clean up data between tests
- Use unique identifiers

### 2. Test Data
- Use predictable test data
- Include edge cases
- Test both success and failure scenarios

### 3. Assertions
- Use descriptive assertion messages
- Test both positive and negative cases
- Verify side effects

### 4. Performance
- Use `testing.Short()` to skip slow tests
- Consider test timeouts
- Parallel execution where possible

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Run unit tests
        run: go test ./... -short -v
      - name: Run integration tests
        run: go test ./... -v
        env:
          CASSANDRA_HOST: localhost
          CASSANDRA_PORT: 9042
```

## Debugging Tests

### Enable Query Logging
```go
const testConfig = `
cassandra:
  hosts: ["localhost:9042"]
  keyspace: "myjournal"
  log_queries: true  # Enable this for debugging
`
```

### Common Issues
1. **Connection refused**: Cassandra not running
2. **Keyspace not found**: Schema not created
3. **Timeout**: Cassandra not ready yet
4. **UUID errors**: Type conversion issues

## Recommendations

1. **Start with integration tests** for new features
2. **Add testcontainers** for CI/CD
3. **Consider unit tests** for complex business logic
4. **Use test tags** to control test execution
5. **Keep tests fast** and reliable
