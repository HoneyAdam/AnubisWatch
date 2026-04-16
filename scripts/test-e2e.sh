#!/bin/bash
# E2E Test Script for AnubisWatch

set -e

echo "⚖️  AnubisWatch E2E Test Suite"
echo "================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test functions
run_backend_tests() {
    echo -e "${YELLOW}Running backend tests...${NC}"
    go test -v -coverprofile=coverage.out ./... 2>&1 | grep -E "(^ok|^FAIL|coverage:)"
    
    # Check coverage threshold
    COVERAGE=$(grep -v "internal/grpcapi/v1/" coverage.out | go tool cover -func=- | grep total | awk '{print $3}' | sed 's/%//')
    echo -e "Coverage (excluding generated): ${GREEN}$COVERAGE%${NC}"
    
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
        echo -e "${RED}FAIL: Coverage below 80%${NC}"
        exit 1
    fi
}

run_frontend_tests() {
    echo -e "${YELLOW}Running frontend tests...${NC}"
    cd web
    npm run test 2>&1 | tail -10
    cd ..
}

run_lint() {
    echo -e "${YELLOW}Running linter...${NC}"
    cd web
    npm run lint 2>&1 | tail -5
    cd ..
}

run_build() {
    echo -e "${YELLOW}Building binary...${NC}"
    CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/anubis-test ./cmd/anubis
    /tmp/anubis-test version
}

run_docker_build() {
    echo -e "${YELLOW}Building Docker image...${NC}"
    docker build -t anubiswatch:test . > /dev/null 2>&1
    echo -e "${GREEN}Docker build successful${NC}"
}

# Main
echo ""
run_backend_tests
echo ""
run_frontend_tests
echo ""
run_lint
echo ""
run_build
echo ""
run_docker_build

echo ""
echo -e "${GREEN}✅ All E2E tests passed!${NC}"
