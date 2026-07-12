# --- Build stage -------------------------------------------------------------
# lola-core has no cgo dependency (modernc.org/sqlite is pure Go), so the
# build stage doesn't need gcc/musl-dev — just the Go toolchain.
FROM golang:1.22-bookworm AS build

WORKDIR /src

# The compose build context is the repo root (lola/), so lola-core lives at
# lola-os/lola-core. Cache module downloads separately from source changes.
COPY lola-os/lola-core/go.mod lola-os/lola-core/go.sum* ./
RUN go mod download

COPY lola-os/lola-core/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=docker" \
    -o /out/lola ./cmd/lola

# --- Runtime stage -------------------------------------------------------------
# distroless gives us a minimal, low-CVE-surface image with no shell —
# appropriate for a binary that handles private key material.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/lola /usr/local/bin/lola

# lola-core writes its vault, registry, and config under ~/.lola. The
# distroless nonroot user's HOME is /home/nonroot.
ENV HOME=/home/nonroot
VOLUME ["/home/nonroot/.lola"]

# 8899: JSON-RPC TCP listener, used by SDKs/test harnesses connecting
# from outside the container.
#
# Port 8765 (the HITL WebSocket channel) is deliberately NOT exposed
# here: lola-core refuses to bind that server to anything but a loopback
# address (see internal/hitl/websocket.go) since the protocol has no
# authentication. docker-compose.yml's optional dashboard sidecar reaches
# it via `network_mode: "service:lola-core"` (sharing the container's
# network namespace) rather than a host port mapping — see that file's
# comments for why.
EXPOSE 8899

ENTRYPOINT ["/usr/local/bin/lola"]
CMD ["serve", "--tcp", "0.0.0.0:8899"]
