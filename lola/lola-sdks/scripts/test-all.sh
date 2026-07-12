#!/usr/bin/env bash
# Runs tests for all three LOLA OS SDKs. Run from lola-sdks/.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> Python SDK"
(cd python && python3 -m pytest)

echo "==> Go SDK"
(cd go && go test ./...)

echo "==> TypeScript SDK"
(cd typescript && npm test)

echo "All SDK test suites passed."
