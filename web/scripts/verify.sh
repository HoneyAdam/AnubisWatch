#!/bin/bash
# Full System Verification Script

echo "═══════════════════════════════════════════════════"
echo "⚖️  AnubisWatch Full System Verification"
echo "═══════════════════════════════════════════════════"
echo ""

PASS=0
FAIL=0

check() {
    if [ $1 -eq 0 ]; then
        echo "✓ $2"
        PASS=$((PASS+1))
    else
        echo "✗ $2"
        FAIL=$((FAIL+1))
    fi
}

echo "1. Build Tests"
echo "───────────────────────────────────────────────────"
CGO_ENABLED=0 go build -o /tmp/anubis-verify ./cmd/anubis 2>/dev/null
check $? "Go binary build"

cd web && npm run build 2>/dev/null && cd ..
check $? "Frontend build"

echo ""
echo "2. Unit Tests"
echo "───────────────────────────────────────────────────"
go test -short ./... 2>/dev/null | grep -q "FAIL"
if [ $? -eq 0 ]; then
    check 1 "Go unit tests"
else
    check 0 "Go unit tests"
fi

cd web && npm run test 2>/dev/null | grep -q "passed" && cd ..
check $? "Frontend unit tests"

echo ""
echo "3. Deployment Files"
echo "───────────────────────────────────────────────────"
[ -f Dockerfile ] && check 0 "Dockerfile exists" || check 1 "Dockerfile exists"
[ -f docker-compose.yml ] && check 0 "docker-compose.yml exists" || check 1 "docker-compose.yml exists"
[ -d deploy/k8s ] && check 0 "K8s manifests exist" || check 1 "K8s manifests exist"
[ -f deploy/helm/anubiswatch/Chart.yaml ] && check 0 "Helm chart exists" || check 1 "Helm chart exists"
[ -f .github/workflows/ci.yml ] && check 0 "CI workflow exists" || check 1 "CI workflow exists"

echo ""
echo "═══════════════════════════════════════════════════"
echo "Results: $PASS passed, $FAIL failed"
echo "═══════════════════════════════════════════════════"

if [ $FAIL -gt 0 ]; then
    exit 1
else
    echo "✓ All systems operational"
    exit 0
fi
