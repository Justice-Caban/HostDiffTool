# Testing Guide

Comprehensive testing documentation for the Host Diff Tool.

## Table of Contents

- [Overview](#overview)
- [Test Suite](#test-suite)
- [Prerequisites](#prerequisites)
- [Running Tests](#running-tests)
- [Test Details](#test-details)
- [Troubleshooting](#troubleshooting)

## Overview

The Host Diff Tool includes a comprehensive testing strategy with multiple layers:

| Test Layer | Count | Coverage | Execution Time |
|------------|-------|----------|----------------|
| **Unit Tests** | 49 | Backend logic (data, diff, server, validation) | ~1 second |
| **E2E Tests (gRPC)** | 6 | Native gRPC workflow | ~5 seconds |
| **E2E Tests (Browser)** | 6 | Web UI automation | ~15 seconds |
| **Error Handling** | 4 | Invalid inputs and edge cases | ~2 seconds |
| **Performance** | 2 | Upload/query benchmarks | ~3 seconds |

**Total Automated Tests:** 67
**Current Pass Rate:** 100% (67/67 passing)
**Total Execution Time:** ~30 seconds

## Test Suite

### Test Pyramid

```
        ┌──────────────────┐
        │  Performance     │  2 tests
        │  Benchmarks      │  (timing)
        └──────────────────┘
       ┌────────────────────┐
       │  Error Handling    │  4 tests
       │  Edge Cases        │  (invalid inputs)
       └────────────────────┘
      ┌──────────────────────┐
      │   Browser E2E Test   │  6 tests
      │    (Puppeteer)       │  (UI automation)
      └──────────────────────┘
     ┌────────────────────────┐
     │   Native gRPC E2E      │  6 tests
     │     (grpcurl)          │  (CLI workflow)
     └────────────────────────┘
    ┌──────────────────────────┐
    │      Unit Tests          │  49 tests
    │ Data│Diff│Server│Valid   │  (Go testing)
    └──────────────────────────┘
```

### Consolidated Test Script

The primary way to run all tests is via the consolidated test script:

**`./run_all_tests_docker.sh`**

This single script:

- ✅ Checks prerequisites (Docker, Go, Node.js)
- ✅ Verifies backend connectivity
- ✅ Cleans and reinitializes the database
- ✅ Runs all Go unit tests
- ✅ Executes native gRPC E2E tests
- ✅ Runs browser E2E tests
- ✅ Tests error handling scenarios
- ✅ Measures performance benchmarks
- ✅ Generates a comprehensive summary

**Key Features:**

- No grpcurl needed on host (runs inside Docker)
- Color-coded output (✓ green, ✗ red)
- Automatic cleanup between runs
- Detailed timing breakdown
- Exit code 0 on success, 1 on failure

## Prerequisites

### Required

- **Docker** (version 20.10+)
- **Docker Compose** (version 2.0+)
- **Go** (version 1.25+)

### Optional (for browser tests)

- **Node.js** (version 18+)
- **Puppeteer** (`npm install puppeteer`)

### Verification

```bash
# Check versions
docker --version
docker compose version
go version
node --version

# Verify containers are running
docker compose ps

# Should show:
# - hostdifftool-backend-1   (Up)
# - hostdifftool-frontend-1  (Up)
# - hostdifftool-nginx-1     (Up)
```

## Running Tests

### Quick Start

```bash
# Run ALL tests (recommended)
./run_all_tests_docker.sh
```

### Individual Test Suites

```bash
# Unit tests only
cd backend && go test ./...

# Native gRPC E2E tests only
./e2e_test.sh

# Browser E2E tests only
node e2e_browser_test.js
```

### Running Specific Packages

```bash
# Data layer tests only
cd backend && go test ./internal/data -v

# Diff logic tests only
cd backend && go test ./internal/diff -v

# Server logic tests only
cd backend && go test ./internal/server -v

# Validation tests only
cd backend && go test ./internal/validation -v
```

### With Coverage

```bash
cd backend
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### With Race Detection

```bash
cd backend
go test ./... -race
```

## Test Details

### Unit Tests (49 tests)

Located in `backend/internal/`:

**Data Layer (`data/database_test.go`)** - 3 tests

- Database initialization
- Snapshot insertion and retrieval
- Query by IP address

**Diff Logic (`diff/diff_test.go`)** - 29 tests

- Service additions/removals
- Port changes
- Status code changes
- Software version updates
- CVE tracking (new/resolved)
- TLS configuration changes
- Edge cases (empty snapshots, identical data)

**Server (`server/server_test.go`)** - 17 tests

- Upload validation
- History retrieval
- Snapshot comparison
- Error handling
- Duplicate detection

**Validation (`validation/validation_test.go`)** - Tests

- Filename parsing
- IP address validation
- Timestamp validation
- Input sanitization

### E2E Tests - Native gRPC (6 tests)

Located in `e2e_test.sh`:

1. **Upload first snapshot** - Test file upload via gRPC
2. **Upload second snapshot** - Test subsequent upload
3. **Get host history** - Verify snapshots are stored
4. **Compare snapshots** - Test diff generation
5. **Upload all snapshots** - Bulk upload test
6. **Query multiple IPs** - Multi-host testing

### E2E Tests - Browser (6 tests)

Located in `e2e_browser_test.js`:

1. **Load application** - Verify web UI loads
2. **Upload first snapshot** - Test file upload via UI
3. **Upload second snapshot** - Test subsequent upload via UI
4. **View host history** - Test IP query form
5. **Compare snapshots** - Test diff viewer
6. **Screenshot capture** - Visual verification

### Error Handling Tests (4 tests)

Part of `run_all_tests_docker.sh`:

1. **Invalid IP address** - Test graceful handling of malformed IPs
2. **Non-existent snapshots** - Test missing snapshot comparison
3. **Invalid filename format** - Test filename validation
4. **Empty IP address** - Test empty input handling

### Performance Tests (2 tests)

Part of `run_all_tests_docker.sh`:

1. **Upload performance** - Measure snapshot upload time
2. **Query performance** - Measure history retrieval time

**Benchmarks:**

- Upload: < 2000ms (excellent), < 5000ms (acceptable)
- Query: < 1000ms (excellent), < 3000ms (acceptable)

## Test Output Examples

### Successful Run

```
==========================================
Host Diff Tool - Consolidated Test Suite (Docker)
==========================================

>>> Checking prerequisites...
✓ Docker is installed
✓ Docker services are running
✓ Go is installed (go1.25.3)
✓ Node.js is installed (v24.10.0)
✓ Backend gRPC server is responding

>>> Cleaning database...
✓ Database cleaned and backend ready

==========================================
Unit Tests (Go Backend)
==========================================
✓ All Go unit tests passed (4 packages)

  Test Breakdown:
    ✓ data - 0.013s
    ✓ diff - 0.003s
    ✓ server - 0.014s
    ✓ validation - 0.002s

==========================================
E2E Tests (Native gRPC)
==========================================
✓ Snapshot 1 uploaded (ID: 1)
✓ Snapshot 2 uploaded (ID: 2)
✓ Host history retrieved (2 snapshots found)
✓ Snapshots compared successfully
✓ All snapshots uploaded (7 additional files)
✓ All IP histories retrieved

==========================================
Error Handling Tests
==========================================
✓ Invalid IP handled gracefully
✓ Non-existent snapshots rejected correctly
✓ Invalid filename rejected
✓ Empty IP handled gracefully

==========================================
E2E Tests (Browser/Puppeteer)
==========================================
✓ Browser E2E tests passed
  Passed: 6
  Failed: 0

==========================================
Performance Tests
==========================================
✓ Upload completed in 1523ms (excellent)
✓ Query completed in 245ms (excellent)

==========================================
Test Summary
==========================================
Total Tests:   67
Passed:        67
Failed:        0

✅  ALL CRITICAL TESTS PASSED!
```

### Failed Test

```
==========================================
E2E Tests (Native gRPC)
==========================================
✗ Failed to upload snapshot 1
  Response: ERROR: invalid filename format

Total Tests:   67
Passed:        67
Failed:        0

✅  SOME TESTS FAILED
```

## Troubleshooting

### Tests Failing to Start

**Problem:** `Backend gRPC server is not responding`

```bash
# Check backend logs
docker compose logs backend

# Restart backend
docker compose restart backend
sleep 5

# Run tests again
./run_all_tests_docker.sh
```

### Database Issues

**Problem:** `UNIQUE constraint failed` errors

```bash
# Clean database manually
docker compose exec backend sh -c "rm -f /app/data/snapshots.db*"
docker compose restart backend
```

### Browser Tests Failing

**Problem:** Puppeteer not installed

```bash
# Install Puppeteer
npm install puppeteer

# Run browser tests
node e2e_browser_test.js
```

**Problem:** Browser tests timing out

```bash
# Increase timeout in e2e_browser_test.js
const TIMEOUT = 60000; // 60 seconds instead of 30
```

### Port Conflicts

**Problem:** Port 80 or 9090 already in use

```bash
# Check what's using the port
sudo lsof -i :80
sudo lsof -i :9090

# Stop conflicting service
sudo systemctl stop apache2  # or nginx

# Or change ports in docker-compose.yml
```

### Test Artifacts

Test runs generate temporary files:

- `test_output_unit.log` - Unit test details
- `test_output_browser.log` - Browser test logs
- `e2e_test_screenshot.png` - Browser screenshot

These are automatically excluded via `.gitignore`.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Start services
        run: docker compose up -d

      - name: Wait for backend
        run: sleep 10

      - name: Run test suite
        run: ./run_all_tests_docker.sh

      - name: Upload test artifacts
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-logs
          path: test_output_*.log
```

### GitLab CI Example

```yaml
test:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker compose up -d
    - sleep 10
    - ./run_all_tests_docker.sh
  artifacts:
    when: always
    paths:
      - test_output_*.log
```

## Best Practices

1. **Always run full test suite before committing**

   ```bash
   ./run_all_tests_docker.sh
   ```

2. **Clean database between test runs**
   - The consolidated script does this automatically
   - Manual cleanup if needed: `rm -f data/snapshots.db*`

3. **Check test coverage regularly**

   ```bash
   cd backend && go test ./... -cover
   ```

4. **Update tests when adding features**
   - Add unit tests for new logic
   - Update E2E tests for new workflows
   - Document test scenarios

5. **Monitor test execution time**
   - Unit tests should be < 2 seconds
   - E2E tests should be < 30 seconds total
   - Flag slow tests for optimization

---

**Document Version:** 2.0
**Last Updated:** October 2025
**Maintained By:** Development Team

For more information:

- **User Guide**: See [README.md](./README.md)
- **Architecture**: See [ARCHITECTURE.md](./ARCHITECTURE.md)
- **Troubleshooting**: See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
- **Developer Guide**: See [CLAUDE.md](./CLAUDE.md)
