#!/usr/bin/env bash
# Builds and starts lola-core via docker-compose, then runs an example
# replay plan inside the running container.
#
# Usage:
#   ./scripts/run-replay.sh [plan-file]
#
# plan-file defaults to examples/plan.json (mounted into the container at
# /examples/plan.json).
set -euo pipefail
cd "$(dirname "$0")/.."

PLAN_FILE="${1:-plan.json}"
CONTAINER_PLAN_PATH="/examples/$(basename "$PLAN_FILE")"

echo "==> Building and starting lola-core"
docker compose up -d --build lola-core

echo "==> Waiting for lola-core to become healthy"
for _ in $(seq 1 30); do
  status="$(docker inspect --format='{{.State.Health.Status}}' lola-core 2>/dev/null || echo "starting")"
  if [ "$status" = "healthy" ]; then
    break
  fi
  sleep 2
done

echo "==> Running replay plan: $CONTAINER_PLAN_PATH (dry-run)"
docker compose exec -T lola-core lola replay "$CONTAINER_PLAN_PATH" --dry-run

echo
echo "Done. To run for real (not dry-run), set up a vault key first:"
echo "  docker compose exec lola-core lola vault init"
echo "  docker compose exec lola-core lola vault add deployer"
echo "  docker compose exec lola-core lola replay $CONTAINER_PLAN_PATH --key-name deployer"
