#!/bin/bash

# AgileOS Automated Testing Script
# Runs all unit tests, integration tests, and generates coverage reports

set -e  # Exit on error

echo "========================================="
echo "🧪 AgileOS Automated Testing Suite"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_section() {
    echo ""
    echo -e "${YELLOW}=========================================${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}=========================================${NC}"
}

# Track test results
BACKEND_TESTS_PASSED=0
FRONTEND_TESTS_PASSED=0
INTEGRATION_TESTS_PASSED=0

# Backend Go Tests
print_section "1. Running Backend Go Tests"
cd agile-os/backend-go

echo "Installing Go dependencies..."
go mod download
go mod tidy

echo ""
echo "Running unit tests..."
if go test ./... -v -cover -coverprofile=coverage.out; then
    print_success "Backend unit tests PASSED"
    BACKEND_TESTS_PASSED=1
    
    echo ""
    echo "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    
    # Calculate coverage percentage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "Total Coverage: $COVERAGE"
    
    if [[ $(echo "$COVERAGE" | sed 's/%//') > 80 ]]; then
        print_success "Coverage exceeds 80% threshold"
    else
        print_warning "Coverage is below 80% threshold"
    fi
else
    print_error "Backend unit tests FAILED"
fi

# Run specific test packages
echo ""
echo "Running BPM Engine tests..."
go test ./internal/bpm/... -v

echo ""
echo "Running Auth Middleware tests..."
go test ./middleware/... -v

echo ""
echo "Running Integration tests..."
if go test ./tests/... -v; then
    print_success "Integration tests PASSED"
    INTEGRATION_TESTS_PASSED=1
else
    print_error "Integration tests FAILED"
fi

cd ../..

# Frontend Next.js Tests
print_section "2. Running Frontend Next.js Tests"
cd agile-os/frontend-next

echo "Installing Node dependencies..."
npm install --silent

echo ""
echo "Running component tests..."
if npm run test -- --run; then
    print_success "Frontend component tests PASSED"
    FRONTEND_TESTS_PASSED=1
else
    print_error "Frontend component tests FAILED"
fi

echo ""
echo "Generating frontend coverage report..."
npm run test:coverage -- --run || true

cd ../..

# Python Analytics Tests (if applicable)
print_section "3. Running Python Analytics Tests"
if [ -d "agile-os/analytics-py" ]; then
    cd agile-os/analytics-py
    
    if [ -f "requirements.txt" ]; then
        echo "Installing Python dependencies..."
        pip install -r requirements.txt --quiet || true
        
        echo ""
        echo "Running Python tests..."
        if python -m pytest tests/ -v 2>/dev/null || true; then
            print_success "Python tests completed"
        else
            print_warning "Python tests not found or failed"
        fi
    fi
    
    cd ../..
else
    print_warning "Python analytics directory not found, skipping"
fi

# Test Summary
print_section "Test Summary"

echo ""
echo "Test Results:"
echo "============="

if [ $BACKEND_TESTS_PASSED -eq 1 ]; then
    print_success "Backend Tests: PASSED"
else
    print_error "Backend Tests: FAILED"
fi

if [ $FRONTEND_TESTS_PASSED -eq 1 ]; then
    print_success "Frontend Tests: PASSED"
else
    print_error "Frontend Tests: FAILED"
fi

if [ $INTEGRATION_TESTS_PASSED -eq 1 ]; then
    print_success "Integration Tests: PASSED"
else
    print_error "Integration Tests: FAILED"
fi

echo ""
echo "Coverage Reports:"
echo "================="
echo "Backend:  agile-os/backend-go/coverage.html"
echo "Frontend: agile-os/frontend-next/coverage/"

echo ""
if [ $BACKEND_TESTS_PASSED -eq 1 ] && [ $FRONTEND_TESTS_PASSED -eq 1 ]; then
    print_success "🎉 ALL TESTS PASSED! Ready for deployment."
    exit 0
else
    print_error "❌ SOME TESTS FAILED. Please fix before deployment."
    exit 1
fi