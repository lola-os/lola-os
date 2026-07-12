# lola-infra

Docker Compose test environment, deployment scripts, and example
`plan.json` files for LOLA OS.

The multi-stage Dockerfile compiles `lola-core` (the same `go build` the
engine's own test suite runs) into a distroless runtime image. The build
context is the monorepo root, so the compose file references
`lola-os/lola-core` for the source and `lola-infra/` for the compose,
config, and dashboard assets.

## What's here

```
lola-infra/
├── Dockerfile               Multi-stage build: Go → distroless runtime
├── docker-compose.yml         lola-core + optional dashboard demo
├── config.docker.yaml         Pre-configured config.yaml for the container (enables HITL WebSocket)
├── .env.example                Copy to .env before running
├── dashboard/                 Minimal HITL approval UI demo (nginx + static HTML/JS)
│   ├── index.html
│   └── nginx.conf
├── examples/
│   └── plan.json               Example structured execution plan
└── scripts/
    ├── run-replay.sh            Build, start, and run a replay plan
    ├── integration-test.sh       Smoke test: doctor, registry, replay, metrics
    └── deploy-systemd.sh         Alternative: run lola-core as a systemd service, no Docker
```

## Quick start

```bash
cd lola-infra
cp .env.example .env   # edit LOLA_VAULT_PASSPHRASE
docker compose up -d --build lola-core
docker compose exec lola-core lola doctor
```

Run the example replay plan (safe — dry-run by default):

```bash
./scripts/run-replay.sh
```

or directly:

```bash
docker compose exec lola-core lola replay /examples/plan.json --dry-run
```

## The optional HITL dashboard demo

```bash
docker compose --profile dashboard up -d
```

Then open `http://localhost:8080` — you'll see a minimal grayscale UI
that connects to lola-core's WebSocket approval channel and renders
incoming approval requests with Approve/Deny/Skip buttons.

### Why this needs a reverse proxy

lola-core's WebSocket HITL server refuses to bind to anything but a
loopback address (`127.0.0.1`) — there's no authentication on that
protocol, so binding it to a network-reachable address would let
*anyone* on the network approve or deny your transactions. That's correct
behavior and this compose setup doesn't change it.

The complication: a browser running on your host machine can't reach a
container's internal loopback address directly, even when two containers
share a network namespace (`network_mode: "service:lola-core"`, used
here) — that trick lets one *container's process* reach another
container's loopback, but your browser isn't a process inside that
namespace.

The fix is `dashboard/nginx.conf`: it runs *inside* lola-core's shared
namespace, so its own connection to `127.0.0.1:8765` is really
lola-core's loopback-only WebSocket server — and it reverse-proxies that
out to `:8080/ws`, which **is** exposed to your host (via the port
mapping on the `lola-core` service in `docker-compose.yml`, which applies
to the whole shared namespace). lola-core's bind address never changes;
only the dashboard's own proxy is reachable externally.

## Running tests

```bash
./scripts/integration-test.sh
```

This builds the image, starts the container, runs `lola doctor`,
exercises `lola registry` and `lola metrics`, and runs a dry-run replay
— then tears everything down. Note: RPC connectivity checks inside
`lola doctor` will fail in a fully offline CI runner (no route to public
RPC endpoints) — that's expected and the script treats it as a soft
warning rather than a hard failure, since the config/vault/registry
checks are what this particular harness can meaningfully verify without
real network access.

## Non-Docker deployment

If you'd rather run `lola-core` directly on a Linux host:

```bash
sudo ./scripts/deploy-systemd.sh /path/to/built/lola-core/bin/lola
```

This installs the binary, creates a dedicated `lola` system user, and
writes a systemd unit that reads the vault passphrase from
`/etc/lola-core/lola-core.env` (never from `systemctl status` output or
shell history).

## Persistent data

The vault and registry live in a named Docker volume (`lola-data`), so
they survive `docker compose down` (but not `docker compose down -v`,
which removes volumes too — `integration-test.sh` does this
deliberately, since it's meant to leave no trace).
