# AnubisWatch E2E Tests

End-to-end tests using Playwright.

## Setup

```bash
cd e2e
npm install
npx playwright install
```

## Run Tests

```bash
# Run all tests
npm test

# Run with UI mode
npm run test:ui

# Run in debug mode
npm run test:debug

# View report
npm run test:report
```

## Environment Variables

- `ANUBIS_URL`: Base URL for tests (default: http://localhost:8080)
- `CI`: Set for CI mode with retries and screenshots

## Test Structure

- `tests/dashboard.spec.ts`: Dashboard navigation tests
- `tests/auth.spec.ts`: Authentication tests
- `tests/api.spec.ts`: API endpoint tests
