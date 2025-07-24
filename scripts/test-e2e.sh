#!/bin/bash

# Simple End-to-End Test - Uses organized test directory structure
# This test creates all files in tests/ directory to keep root clean

set -e  # Exit on any error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test directories
TEST_DIR="tests/e2e"
TEST_OUTPUT_DIR="tests/output"

echo -e "${BLUE}🚀 Simple End-to-End Test${NC}"
echo "=================================="
echo "Project root: $(pwd)"
echo "Test directory: $TEST_DIR"
echo "Output directory: $TEST_OUTPUT_DIR"

# Setup test environment
echo -e "\n${BLUE}🏗️  Setting up test environment${NC}"

# Create test directories
mkdir -p "$TEST_DIR" "$TEST_OUTPUT_DIR"

# Store original directory
ORIGINAL_DIR="$(pwd)"

# Change to test directory
cd "$TEST_DIR"

echo "Working in: $(pwd)"

# Cleanup function - but ask before cleaning
cleanup() {
    echo -e "\n${YELLOW}🧹 Test completed. Clean up test files? (y/n)${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        cd "$ORIGINAL_DIR"
        rm -rf tests/
        echo "✅ Test files cleaned up"
    else
        cd "$ORIGINAL_DIR"
        echo -e "${GREEN}📁 Test files preserved in tests/ directory${NC}"
        echo "- Test sources: $TEST_DIR/"
        echo "- Test outputs: $TEST_OUTPUT_DIR/"
    fi
}

# Set up cleanup trap - but only on successful completion
trap cleanup EXIT

# Step 1: Check prerequisites
echo -e "\n${BLUE}📋 Step 1: Checking prerequisites${NC}"

if [ ! -f "../../bin/pm" ]; then
    echo -e "${RED}❌ Binary not found. Building...${NC}"
    (cd ../.. && go build -o bin/pm ./cmd/pm) || { echo "Build failed"; exit 1; }
fi

# Test SSH connection
if ! ssh -o BatchMode=yes -o ConnectTimeout=5 localhost exit 2>/dev/null; then
    echo -e "${RED}❌ SSH connection to localhost failed${NC}"
    echo "Please set up SSH key authentication to localhost"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites met${NC}"

# Step 2: Create test files in test directory
echo -e "\n${BLUE}📁 Step 2: Creating test files${NC}"

# Create some test files you can see
mkdir -p test-src test-docs
echo 'package main

import "fmt"

func main() {
    fmt.Println("Hello from E2E test!")
}' > test-src/main.go

echo 'package utils

// Helper provides utility functions
func Helper() string {
    return "helper function from E2E test"
}' > test-src/utils.go

echo '# Test Documentation

This is a test file for E2E testing.

## Features
- Package creation
- Package upload  
- Package download
- Package extraction

Created by: E2E Test Suite' > test-docs/README.md

# Create files that should be excluded
echo "This should be excluded from the package" > test-temp.tmp
echo "Debug log entry from E2E test" > test-debug.log

echo "Created test files in $(pwd):"
find . -name "test-*" -o -name "*.tmp" -o -name "*.log" | sort

# Step 3: Create packet configuration
echo -e "\n${BLUE}📦 Step 3: Creating packet configuration${NC}"

cat > test-e2e-packet.json << 'EOF'
{
  "name": "e2e-simple-test",
  "ver": "1.0.0",
  "targets": [
    {
      "path": "test-src/*.go",
      "exclude": []
    },
    "test-docs/*.md",
    {
      "path": "*",
      "exclude": ["*.tmp", "*.log", "*.json"]
    }
  ]
}
EOF

echo "Packet configuration created:"
cat test-e2e-packet.json

# Step 4: Create SSH configuration pointing to test output
echo -e "\n${BLUE}🔐 Step 4: Creating SSH configuration${NC}"

TEST_REMOTE_DIR="$ORIGINAL_DIR/$TEST_OUTPUT_DIR/remote-packages"
mkdir -p "$TEST_REMOTE_DIR"

cat > ssh-config.json << EOF
{
  "host": "localhost",
  "port": 22,
  "username": "$(whoami)",
  "key_path": "~/.ssh/id_rsa",
  "remote_dir": "$TEST_REMOTE_DIR"
}
EOF

echo "SSH config points to: $TEST_REMOTE_DIR"

# Step 5: Test create command
echo -e "\n${BLUE}📤 Step 5: Testing package creation${NC}"

echo "Running: ../../bin/pm create test-e2e-packet.json -c ssh-config.json"
if ../../bin/pm create test-e2e-packet.json -c ssh-config.json; then
    echo -e "${GREEN}✅ Package created successfully${NC}"
else
    echo -e "${RED}❌ Package creation failed${NC}"
    exit 1
fi

# Step 6: Check if package was uploaded
echo -e "\n${BLUE}🔍 Step 6: Verifying upload${NC}"

ARCHIVE_FILE="$TEST_REMOTE_DIR/e2e-simple-test-1.0.0.tar.gz"
if [ -f "$ARCHIVE_FILE" ]; then
    echo -e "${GREEN}✅ Package found: $ARCHIVE_FILE${NC}"
    echo "Archive size: $(ls -lh "$ARCHIVE_FILE" | awk '{print $5}')"
    
    # Show archive contents for verification
    echo "Archive contents:"
    tar -tzf "$ARCHIVE_FILE" | head -10
    
    # Save archive info to output
    echo "Archive created at: $(date)" > "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/archive-info.txt"
    echo "Archive file: $ARCHIVE_FILE" >> "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/archive-info.txt"
    echo "Archive size: $(ls -lh "$ARCHIVE_FILE" | awk '{print $5}')" >> "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/archive-info.txt"
    echo "Archive contents:" >> "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/archive-info.txt"
    tar -tzf "$ARCHIVE_FILE" >> "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/archive-info.txt"
else
    echo -e "${RED}❌ Package not found${NC}"
    echo "Contents of $TEST_REMOTE_DIR:"
    ls -la "$TEST_REMOTE_DIR" || echo "Directory doesn't exist"
    exit 1
fi

# Step 7: Create packages configuration
echo -e "\n${BLUE}📥 Step 7: Creating packages configuration${NC}"

cat > test-e2e-packages.json << 'EOF'
{
  "packages": [
    {
      "name": "e2e-simple-test",
      "ver": ">=1.0.0"
    }
  ]
}
EOF

echo "Packages configuration created:"
cat test-e2e-packages.json

# Step 8: Test update command
echo -e "\n${BLUE}📥 Step 8: Testing package download${NC}"

echo "Running: ../../bin/pm update test-e2e-packages.json -c ssh-config.json"
if ../../bin/pm update test-e2e-packages.json -c ssh-config.json; then
    echo -e "${GREEN}✅ Package downloaded successfully${NC}"
else
    echo -e "${RED}❌ Package download failed${NC}"
    exit 1
fi

# Step 9: Verify extracted files
echo -e "\n${BLUE}🔍 Step 9: Verifying extracted content${NC}"

if [ ! -d "packages/e2e-simple-test" ]; then
    echo -e "${RED}❌ Package directory not created${NC}"
    echo "Contents of packages/:"
    ls -la packages/ || echo "packages/ directory doesn't exist"
    exit 1
fi

echo "Contents of packages/e2e-simple-test/:"
find packages/e2e-simple-test -type f | sort

# Check specific files
expected_files=(
    "packages/e2e-simple-test/test-src/main.go"
    "packages/e2e-simple-test/test-src/utils.go"
    "packages/e2e-simple-test/test-docs/README.md"
)

echo -e "\nChecking expected files:"
for file in "${expected_files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ Found: $file${NC}"
        # Show first line of content
        echo "   Content preview: $(head -n 1 "$file")"
    else
        echo -e "${RED}❌ Missing: $file${NC}"
        exit 1
    fi
done

# Check that excluded files were NOT included
excluded_files=(
    "packages/e2e-simple-test/test-temp.tmp"
    "packages/e2e-simple-test/test-debug.log"
)

echo -e "\nChecking excluded files:"
for file in "${excluded_files[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${RED}❌ Excluded file was included: $file${NC}"
        exit 1
    else
        echo -e "${GREEN}✅ Correctly excluded: $(basename "$file")${NC}"
    fi
done

# Step 10: Save results to output directory
echo -e "\n${BLUE}📦 Step 10: Saving results${NC}"

# Copy extracted packages to output directory
if [ -d "packages" ]; then
    cp -r packages "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/"
    echo "✅ Extracted packages saved to: $TEST_OUTPUT_DIR/packages/"
fi

# Save test source files to output
mkdir -p "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/test-sources"
cp -r test-src test-docs test-*.json "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/test-sources/"
echo "✅ Test source files saved to: $TEST_OUTPUT_DIR/test-sources/"

# Create test summary
cat > "$ORIGINAL_DIR/$TEST_OUTPUT_DIR/test-summary.txt" << EOF
E2E Test Summary
================
Test Date: $(date)
Test Duration: $SECONDS seconds
Working Directory: $(pwd)

Test Results:
✅ Package creation: SUCCESS  
✅ Package upload: SUCCESS
✅ Package download: SUCCESS
✅ Package extraction: SUCCESS
✅ File verification: SUCCESS
✅ Exclude patterns: SUCCESS

Files Tested:
- test-src/main.go ($(wc -l < test-src/main.go) lines)
- test-src/utils.go ($(wc -l < test-src/utils.go) lines)  
- test-docs/README.md ($(wc -l < test-docs/README.md) lines)

Files Excluded:
- test-temp.tmp
- test-debug.log

Archive Location: $TEST_REMOTE_DIR/e2e-simple-test-1.0.0.tar.gz
Extracted Location: $TEST_OUTPUT_DIR/packages/e2e-simple-test/
EOF

echo "✅ Test summary saved to: $TEST_OUTPUT_DIR/test-summary.txt"

# Final success message
echo -e "\n${GREEN}🎉 End-to-End Test Completed Successfully!${NC}"
echo "=============================================="
echo "✅ Package created and uploaded"
echo "✅ Package downloaded and extracted"
echo "✅ Files found and verified"
echo "✅ Exclude patterns working correctly"
echo "✅ Test artifacts saved"

echo -e "\n${YELLOW}📁 Test results saved in:${NC}"
echo "- Test summary: $TEST_OUTPUT_DIR/test-summary.txt"
echo "- Archive info: $TEST_OUTPUT_DIR/archive-info.txt"
echo "- Extracted packages: $TEST_OUTPUT_DIR/packages/"
echo "- Test sources: $TEST_OUTPUT_DIR/test-sources/"
echo "- Remote archives: $TEST_OUTPUT_DIR/remote-packages/"

echo -e "\n${YELLOW}🔍 To inspect results:${NC}"
echo "find $TEST_OUTPUT_DIR -type f | sort"
echo "cat $TEST_OUTPUT_DIR/test-summary.txt"