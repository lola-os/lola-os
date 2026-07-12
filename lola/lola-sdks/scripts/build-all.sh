#!/usr/bin/env bash
# Builds all three LOLA OS SDKs. Run from lola-sdks/.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> Python SDK"
(cd python && pip install -e . --break-system-packages)

echo "==> Go SDK"
(cd go && go mod tidy && go build ./...)

echo "==> TypeScript SDK"
(cd typescript && npm install && npm run build)

echo "All SDKs built."
