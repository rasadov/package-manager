#!/bin/bash

echo "🧪 Running Package Manager Tests"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test directories
TEST_DIR="tests"
TEST_OUTPUT_DIR="tests/output"

# Create test directories
mkdir -p "$TEST_DIR" "$TEST_OUTPUT_DIR"

# Function to run tests with nice output
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "\n${BLUE}📋 Running: $test_name${NC}"
    echo "----------------------------------------"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command"; then
        echo -e "${GREEN}✅ PASSED: $test_name${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}❌ FAILED: $test_name${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Function to print test summary
print_summary() {
    echo -e "\n${BLUE}📊 Test Summary${NC}"
    echo "================================="
    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed!${NC}"
        echo -e "\n${YELLOW}📁 Test artifacts saved in: $TEST_DIR/${NC}"
        return 0
    else
        echo -e "\n${RED}💥 Some tests failed!${NC}"
        return 1
    fi
}

# Cleanup function
cleanup_tests() {
    echo -e "\n${YELLOW}🧹 Clean up test artifacts? (y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        rm -rf "$TEST_DIR"
        echo -e "${GREEN}✅ Test artifacts cleaned up${NC}"
    else
        echo -e "${YELLOW}📁 Test artifacts preserved in $TEST_DIR/${NC}"
    fi
}

# Set cleanup trap
trap cleanup_tests EXIT

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed or not in PATH${NC}"
    exit 1
fi

# Build the project first
echo -e "${YELLOW}🔨 Building project...${NC}"
if ! go build -o bin/pm ./cmd/pm; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# Run all tests together with coverage
run_test "All Unit Tests with Coverage" "go test -cover ./internal/utils ./internal/controller ./config"

# Integration tests (if SSH is available)
if command -v ssh &> /dev/null && ssh -o BatchMode=yes -o ConnectTimeout=2 localhost exit &>/dev/null; then
    echo -e "\n${YELLOW}🔗 SSH available - running integration tests${NC}"
    
    # Make scripts executable
    chmod +x scripts/test-e2e.sh 2>/dev/null || true
    chmod +x test-e2e-simple.sh 2>/dev/null || true
    
    # Try to run integration test
    if [ -f "test-e2e-simple.sh" ]; then
        run_test "End-to-End Workflow Test" "./test-e2e-simple.sh"
    elif [ -f "scripts/test-e2e.sh" ]; then
        run_test "End-to-End Workflow Test" "./scripts/test-e2e.sh"
    else
        echo -e "${YELLOW}⚠️  Integration test script not found${NC}"
    fi
else
    echo -e "\n${YELLOW}⚠️  SSH not available - skipping integration tests${NC}"
    echo "To run integration tests:"
    echo "1. Set up SSH key authentication to localhost"
    echo "2. Run: ./test-e2e-simple.sh"
fi

# Performance/benchmark tests
echo -e "\n${YELLOW}⚡ Running performance tests${NC}"
run_test "Archive Performance Benchmark" "go test -bench=BenchmarkCreateTarGz -run=^$ ./internal/utils 2>/dev/null || go test -bench=. -run=^$ ./internal/utils"

# Race condition tests
echo -e "\n${YELLOW}🏃 Running race condition tests${NC}"
run_test "Race Condition Tests" "go test -race -short ./internal/utils ./config"

# Print final summary
print_summary
exit_code=$?

echo -e "\n${BLUE}🔍 Additional Test Commands:${NC}"
echo "- Run specific test: go test -v ./internal/utils -run TestNameHere"
echo "- Run with coverage report: go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out"
echo "- Run benchmarks: go test -bench=. ./internal/utils"
echo "- Run integration tests: ./test-e2e-simple.sh"
echo "- Inspect test artifacts: find $TEST_DIR -type f"

exit $exit_code