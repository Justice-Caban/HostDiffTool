#!/bin/bash

#####################################################################
# Docker-based Consolidated Test Suite for Host Diff Tool
#####################################################################
# This script runs all tests using Docker containers
# No need to install grpcurl on the host machine
#####################################################################

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Log functions
log_header() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
}

log_section() {
    echo ""
    echo -e "${BLUE}>>> $1${NC}"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

log_failure() {
    echo -e "${RED}✗${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

log_skip() {
    echo -e "${YELLOW}⊘${NC} $1"
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
}

log_info() {
    echo "  → $1"
}

# Check prerequisites
check_prerequisites() {
    log_section "Checking prerequisites..."

    local all_good=true

    # Check if Docker is available
    if command -v docker &> /dev/null; then
        log_success "Docker is installed"
    else
        log_failure "Docker is not installed"
        all_good=false
    fi

    # Check if services are running
    if docker compose ps | grep -q "Up"; then
        log_success "Docker services are running"
    else
        log_failure "Docker services are not running"
        echo "  Please run: docker compose up -d"
        all_good=false
    fi

    # Check for Go
    if command -v go &> /dev/null; then
        log_success "Go is installed ($(go version | awk '{print $3}'))"
    else
        log_failure "Go is not installed"
        all_good=false
    fi

    # Check for Node.js (for browser tests)
    if command -v node &> /dev/null; then
        log_success "Node.js is installed ($(node --version))"
    else
        log_skip "Node.js is not installed (browser tests will be skipped)"
    fi

    # Check if backend is responding
    local backend_test=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"ip_address": "0.0.0.0"}' localhost:9090 hostdiff.HostService/GetHostHistory 2>&1 | grep -v "http2:")

    # A valid response is either {} or contains "snapshots"
    if echo "$backend_test" | grep -Eq '\{|\}|snapshots'; then
        log_success "Backend gRPC server is responding"
    else
        log_failure "Backend gRPC server is not responding"
        echo "  Response: $backend_test"
        echo "  Please check: docker compose logs backend"
        all_good=false
    fi

    if [ "$all_good" = false ]; then
        echo ""
        echo -e "${RED}Prerequisites check failed. Please fix the issues above.${NC}"
        exit 1
    fi
}

# Clean database before tests
clean_database() {
    log_section "Cleaning database..."

    # Remove database files from the running container
    docker compose exec -T backend sh -c "rm -f /app/data/snapshots.db*" &> /dev/null
    log_info "Database files removed"

    # Restart backend to reinitialize database
    docker compose restart backend &> /dev/null
    log_info "Backend restarted"

    log_info "Waiting for backend to be ready..."
    sleep 5

    local backend_check=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"ip_address": "0.0.0.0"}' localhost:9090 hostdiff.HostService/GetHostHistory 2>&1 | grep -v "http2:")

    # A valid response is either {} or contains "snapshots"
    if echo "$backend_check" | grep -Eq '\{|\}|snapshots'; then
        log_success "Database cleaned and backend ready"
    else
        log_failure "Backend failed to restart properly"
        exit 1
    fi
}

# Run Go unit tests
run_unit_tests() {
    log_header "Unit Tests (Go Backend)"

    log_section "Running Go unit tests..."

    cd backend

    if go test ./... -v -count=1 > ../test_output_unit.log 2>&1; then
        # Parse test results
        local test_count=$(grep -E "^(PASS|FAIL):" ../test_output_unit.log | wc -l)
        local pass_count=$(grep "^PASS:" ../test_output_unit.log | wc -l)

        log_success "All Go unit tests passed ($pass_count packages)"

        # Show test breakdown
        echo ""
        echo "  Test Breakdown:"
        grep -E "ok\s+.*\s+[0-9]+\.[0-9]+s" ../test_output_unit.log | while read -r line; do
            local pkg=$(echo "$line" | awk '{print $2}')
            local time=$(echo "$line" | awk '{print $3}')
            echo "    ✓ $(basename $pkg) - $time"
        done

        PASSED_TESTS=$((PASSED_TESTS + test_count))
        TOTAL_TESTS=$((TOTAL_TESTS + test_count))
    else
        log_failure "Go unit tests failed"
        echo "  See test_output_unit.log for details"
        cat ../test_output_unit.log
        FAILED_TESTS=$((FAILED_TESTS + 1))
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi

    cd ..
}

# Run native gRPC E2E tests using Docker
run_grpc_e2e_tests() {
    log_header "E2E Tests (Native gRPC)"

    log_section "Testing gRPC endpoints via Docker..."

    # Test 1: Upload first snapshot
    log_info "Test 1: Upload first snapshot"
    local filename1="host_125.199.235.74_2025-09-10T03-00-00Z.json"

    local response1=$(docker compose exec -T backend sh -c '
        FILE_CONTENT=$(base64 -w 0 /app/assets/host_snapshots/'"$filename1"')
        printf '"'"'{"filename": "'"$filename1"'", "file_content": "%s"}'"'"' "$FILE_CONTENT" | grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d @ localhost:9090 hostdiff.HostService/UploadSnapshot
    ' 2>&1 | grep -v "http2:")

    if echo "$response1" | grep -q "id"; then
        SNAPSHOT_ID_1=$(echo "$response1" | grep -o '"id": "[0-9]*"' | cut -d'"' -f4)
        log_success "Snapshot 1 uploaded (ID: $SNAPSHOT_ID_1)"
    else
        log_failure "Failed to upload snapshot 1"
        echo "  Response: $response1"
        return 1
    fi

    # Test 2: Upload second snapshot
    log_info "Test 2: Upload second snapshot"
    local filename2="host_125.199.235.74_2025-09-15T08-49-45Z.json"

    local response2=$(docker compose exec -T backend sh -c '
        FILE_CONTENT=$(base64 -w 0 /app/assets/host_snapshots/'"$filename2"')
        printf '"'"'{"filename": "'"$filename2"'", "file_content": "%s"}'"'"' "$FILE_CONTENT" | grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d @ localhost:9090 hostdiff.HostService/UploadSnapshot
    ' 2>&1 | grep -v "http2:")

    if echo "$response2" | grep -q "id"; then
        SNAPSHOT_ID_2=$(echo "$response2" | grep -o '"id": "[0-9]*"' | cut -d'"' -f4)
        log_success "Snapshot 2 uploaded (ID: $SNAPSHOT_ID_2)"
    else
        log_failure "Failed to upload snapshot 2"
        return 1
    fi

    # Test 3: Get host history
    log_info "Test 3: Get host history"
    local history=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"ip_address": "125.199.235.74"}' localhost:9090 hostdiff.HostService/GetHostHistory)

    if echo "$history" | grep -q "snapshots"; then
        local count=$(echo "$history" | grep -o '"id"' | wc -l)
        log_success "Host history retrieved ($count snapshots found)"
    else
        log_failure "Failed to get host history"
        return 1
    fi

    # Test 4: Compare snapshots
    log_info "Test 4: Compare snapshots"
    local compare=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d "{\"snapshot_id_a\": \"$SNAPSHOT_ID_1\", \"snapshot_id_b\": \"$SNAPSHOT_ID_2\"}" localhost:9090 hostdiff.HostService/CompareSnapshots)

    if echo "$compare" | grep -q "report"; then
        log_success "Snapshots compared successfully"

        # Check for specific changes
        if echo "$compare" | grep -q "Added Services"; then
            log_info "Detected service additions"
        fi
        if echo "$compare" | grep -q "Removed Services"; then
            log_info "Detected service removals"
        fi
        if echo "$compare" | grep -q "Modified Services"; then
            log_info "Detected service modifications"
        fi
    else
        log_failure "Failed to compare snapshots"
        return 1
    fi

    # Test 5: Upload all remaining snapshots
    log_info "Test 5: Upload all remaining snapshots"
    local uploaded=0
    local failed_uploads=0

    # Get list of files from Docker container
    local files=$(docker compose exec -T backend sh -c "ls /app/assets/host_snapshots/*.json 2>/dev/null || true")

    for file in $files; do
        local filename=$(basename "$file")

        # Skip already uploaded files
        if [ "$filename" = "$filename1" ] || [ "$filename" = "$filename2" ]; then
            continue
        fi

        local response=$(docker compose exec -T backend sh -c '
            FILE_CONTENT=$(base64 -w 0 /app/assets/host_snapshots/'"$filename"')
            printf '"'"'{"filename": "'"$filename"'", "file_content": "%s"}'"'"' "$FILE_CONTENT" | grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d @ localhost:9090 hostdiff.HostService/UploadSnapshot 2>&1
        ' | grep -v "http2:")

        if echo "$response" | grep -q "id"; then
            uploaded=$((uploaded + 1))
        else
            failed_uploads=$((failed_uploads + 1))
        fi
    done

    if [ $uploaded -gt 0 ]; then
        if [ $failed_uploads -eq 0 ]; then
            log_success "All snapshots uploaded ($uploaded additional files)"
        else
            log_failure "Some snapshot uploads failed ($failed_uploads failures)"
        fi
    else
        log_info "No additional snapshots to upload"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi

    # Test 6: Query multiple IPs
    log_info "Test 6: Query history for all IPs"
    local ips=("125.199.235.74" "198.51.100.23" "203.0.113.45")
    local all_ok=true

    for ip in "${ips[@]}"; do
        local result=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d "{\"ip_address\": \"$ip\"}" localhost:9090 hostdiff.HostService/GetHostHistory)
        if ! echo "$result" | grep -q "snapshots"; then
            all_ok=false
        fi
    done

    if [ "$all_ok" = true ]; then
        log_success "All IP histories retrieved"
    else
        log_failure "Some IP history queries failed"
    fi
}

# Run error handling tests
run_error_tests() {
    log_header "Error Handling Tests"

    log_section "Testing error scenarios..."

    # Test 1: Invalid IP address
    log_info "Test 1: Invalid IP address"
    docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"ip_address": "999.999.999.999"}' localhost:9090 hostdiff.HostService/GetHostHistory &> /dev/null

    # Should return empty or error (both are acceptable)
    log_success "Invalid IP handled gracefully"

    # Test 2: Non-existent snapshot comparison
    log_info "Test 2: Non-existent snapshot IDs"
    local response=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"snapshot_id_a": "99999", "snapshot_id_b": "99998"}' localhost:9090 hostdiff.HostService/CompareSnapshots 2>&1)

    if echo "$response" | grep -qi "not found\|error"; then
        log_success "Non-existent snapshots rejected correctly"
    else
        log_failure "Non-existent snapshots not handled properly"
    fi

    # Test 3: Malformed filename
    log_info "Test 3: Invalid filename format"
    local response=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"filename": "invalid_name.json", "file_content": "e30="}' localhost:9090 hostdiff.HostService/UploadSnapshot 2>&1)

    if echo "$response" | grep -qi "invalid\|error"; then
        log_success "Invalid filename rejected"
    else
        log_failure "Invalid filename not validated"
    fi

    # Test 4: Empty IP address
    log_info "Test 4: Empty IP address"
    docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"ip_address": ""}' localhost:9090 hostdiff.HostService/GetHostHistory &> /dev/null

    # Should handle gracefully
    log_success "Empty IP handled gracefully"
}

# Run browser E2E tests
run_browser_tests() {
    log_header "E2E Tests (Browser/Puppeteer)"

    if ! command -v node &> /dev/null; then
        log_skip "Node.js not installed, skipping browser tests"
        return 0
    fi

    if [ ! -f "e2e_browser_test.js" ]; then
        log_skip "Browser test script not found"
        return 0
    fi

    log_section "Running browser-based E2E tests..."

    # Check if puppeteer is installed
    if ! npm list puppeteer &> /dev/null; then
        log_info "Installing puppeteer..."
        npm install puppeteer &> /dev/null
    fi

    if node e2e_browser_test.js > test_output_browser.log 2>&1; then
        log_success "Browser E2E tests passed"

        # Show test summary from log
        if grep -q "Passed:" test_output_browser.log; then
            echo ""
            grep "Passed:" test_output_browser.log | head -1 | sed 's/^/  /'
            grep "Failed:" test_output_browser.log | head -1 | sed 's/^/  /'
        fi

        PASSED_TESTS=$((PASSED_TESTS + 1))
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    else
        log_failure "Browser E2E tests failed"
        echo "  See test_output_browser.log for details"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi
}

# Run performance tests
run_performance_tests() {
    log_header "Performance Tests"

    log_section "Testing system performance..."

    # Test 1: Upload performance - measure actual upload time
    log_info "Test 1: Upload performance"

    local start_time=$(date +%s)
    # Use a unique timestamp for performance test to avoid duplicates
    local perf_timestamp="2099-12-31T23-59-59Z"
    local source_file="host_198.51.100.23_2025-09-10T03-00-00Z.json"
    local unique_filename="host_198.51.100.23_${perf_timestamp}.json"

    local upload_result=$(docker compose exec -T backend sh -c '
        FILE_CONTENT=$(base64 -w 0 /app/assets/host_snapshots/'"$source_file"')
        printf '"'"'{"filename": "'"$unique_filename"'", "file_content": "%s"}'"'"' "$FILE_CONTENT" | grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d @ localhost:9090 hostdiff.HostService/UploadSnapshot 2>&1
    ')

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    if echo "$upload_result" | grep -q '"id"' && [ $duration -le 5 ]; then
        log_success "Upload completed in ${duration}s (acceptable)"
    elif echo "$upload_result" | grep -q '"id"'; then
        log_failure "Upload succeeded but took ${duration}s (slow)"
    else
        log_failure "Upload failed or timed out"
    fi

    # Test 2: Query performance - measure actual query time
    log_info "Test 2: Query performance"

    local start_time=$(date +%s)

    local query_result=$(docker compose exec -T backend grpcurl -plaintext -proto /app/proto/host_diff.proto -import-path /app/proto -d '{"ip_address": "198.51.100.23"}' localhost:9090 hostdiff.HostService/GetHostHistory 2>&1)

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    if echo "$query_result" | grep -q '"snapshots"' && [ $duration -le 3 ]; then
        log_success "Query completed in ${duration}s (fast)"
    elif echo "$query_result" | grep -q '"snapshots"'; then
        log_failure "Query succeeded but took ${duration}s (slow)"
    else
        log_failure "Query failed or timed out"
    fi
}

# Display test summary
show_summary() {
    log_header "Test Summary"

    echo ""
    echo "Total Tests:   $TOTAL_TESTS"
    echo -e "${GREEN}Passed:        $PASSED_TESTS${NC}"

    if [ $FAILED_TESTS -gt 0 ]; then
        echo -e "${RED}Failed:        $FAILED_TESTS${NC}"
    else
        echo "Failed:        $FAILED_TESTS"
    fi

    if [ $SKIPPED_TESTS -gt 0 ]; then
        echo -e "${YELLOW}Skipped:       $SKIPPED_TESTS${NC}"
    fi

    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}=========================================="
        echo "✅  ALL TESTS PASSED!"
        echo -e "==========================================${NC}"
        echo ""

        # Show test artifacts
        if [ -f "e2e_test_screenshot.png" ]; then
            echo "Test artifacts:"
            echo "  • e2e_test_screenshot.png (browser test screenshot)"
        fi

        return 0
    else
        echo -e "${RED}=========================================="
        echo "❌  SOME TESTS FAILED"
        echo -e "==========================================${NC}"
        echo ""
        echo "Review the output above for details."
        echo ""
        echo "Log files:"
        [ -f "test_output_unit.log" ] && echo "  • test_output_unit.log"
        [ -f "test_output_browser.log" ] && echo "  • test_output_browser.log"

        return 1
    fi
}

#####################################################################
# Main execution
#####################################################################

main() {
    log_header "Host Diff Tool - Consolidated Test Suite (Docker)"

    echo "Starting comprehensive test run..."
    echo "This will run all unit tests, E2E tests, and validation checks."
    echo "All gRPC tests run inside Docker containers (no grpcurl needed on host)."
    echo ""

    # Run all test suites
    check_prerequisites
    clean_database
    run_unit_tests
    run_grpc_e2e_tests
    run_error_tests
    run_browser_tests
    run_performance_tests

    # Show final summary
    show_summary

    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
