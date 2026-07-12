# LOLA OS — one place to build, test, and run the whole system.
#
#   make            # show this help
#   make build      # build lola-core (the engine) + all SDKs
#   make test       # run every test suite across the monorepo
#   make doctor     # health-check your local environment
#   make serve      # start the JSON-RPC engine (needs LOLA_VAULT_PASSPHRASE)
#
# Every target is a thin wrapper around the underlying tool, so you can always
# drop down to `go`, `npm`, `pytest`, or `docker compose` directly.

# --- Paths -------------------------------------------------------------------
CORE      := lola-os/lola-core
SDK_PY    := lola-sdks/python
SDK_TS    := lola-sdks/typescript
SDK_GO    := lola-sdks/go
UI_DOCS   := lola-ui
UI_WEB    := ../lolaos-frontend
INFRA     := lola-infra
BIN       := $(CORE)/bin/lola

# Vault passphrase used by targets that need to sign/serve. Override on the CLI:
#   make serve LOLA_VAULT_PASSPHRASE=hunter2
LOLA_VAULT_PASSPHRASE ?=

.DEFAULT_GOAL := help

# --- Meta --------------------------------------------------------------------
.PHONY: help
help: ## Show this help
	@echo "LOLA OS — available targets:"
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
	  | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[1m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build ui-build ## Build everything (engine, SDKs, docs site)

.PHONY: build
build: build-core sdk-py sdk-ts sdk-go ## Build the engine and all three SDKs

.PHONY: test
test: test-core test-py test-ts test-go ## Run every test suite

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BIN)
	rm -rf $(SDK_TS)/dist $(SDK_TS)/dist-test
	rm -rf $(UI_DOCS)/.next $(UI_WEB)/dist

# --- Core engine (Go) --------------------------------------------------------
.PHONY: build-core
build-core: ## Build the lola-core binary
	cd $(CORE) && go mod tidy && go build -o bin/lola ./cmd/lola
	@echo "built $(BIN)"

.PHONY: test-core
test-core: ## Run the lola-core test suite
	cd $(CORE) && go test ./...

.PHONY: vet-core
vet-core: ## go vet the engine
	cd $(CORE) && go vet ./...

.PHONY: serve
serve: build-core ## Start the JSON-RPC engine (stdio + TCP :8899)
	cd $(CORE) && LOLA_VAULT_PASSPHRASE=$(LOLA_VAULT_PASSPHRASE) ./bin/lola serve --tcp 127.0.0.1:8899

.PHONY: doctor
doctor: build-core ## Health-check config, RPC connectivity, vault, registry
	cd $(CORE) && ./bin/lola doctor

.PHONY: chains
chains: build-core ## List every supported chain and which are enabled
	cd $(CORE) && ./bin/lola chains

.PHONY: vault-init
vault-init: build-core ## Create the encrypted vault and add a signing key
	cd $(CORE) && ./bin/lola vault init && ./bin/lola vault add deployer

# --- SDKs --------------------------------------------------------------------
.PHONY: sdk-py
sdk-py: ## Install the Python SDK (editable)
	cd $(SDK_PY) && pip install -e .

.PHONY: test-py
test-py: ## Test the Python SDK
	cd $(SDK_PY) && python3 -m pytest -q

.PHONY: sdk-ts
sdk-ts: ## Build the TypeScript SDK
	cd $(SDK_TS) && npm install && npm run build

.PHONY: test-ts
test-ts: ## Test the TypeScript SDK
	cd $(SDK_TS) && npm test

.PHONY: sdk-go
sdk-go: ## Build the Go SDK (requires lola-core to build first)
	cd $(SDK_GO) && go mod tidy && go build ./...

.PHONY: test-go
test-go: ## Test the Go SDK
	cd $(SDK_GO) && go test ./...

# --- Front ends --------------------------------------------------------------
.PHONY: web
web: ## Run the marketing site dev server (lolaos-frontend)
	cd $(UI_WEB) && npm install && npm run dev

.PHONY: web-build
web-build: ## Production build of the marketing site
	cd $(UI_WEB) && npm install && npm run build

.PHONY: ui
ui: ## Run the docs site dev server (lola-ui)
	cd $(UI_DOCS) && npm install && npm run dev

.PHONY: ui-build
ui-build: ## Production build of the docs site
	cd $(UI_DOCS) && npm install && npm run build

# --- Infra (Docker) ----------------------------------------------------------
.PHONY: infra-up
infra-up: ## Build and start lola-core in Docker
	cd $(INFRA) && docker compose up -d --build lola-core

.PHONY: infra-dashboard
infra-dashboard: ## Start the optional HITL approval dashboard (:8080)
	cd $(INFRA) && docker compose --profile dashboard up -d

.PHONY: infra-down
infra-down: ## Stop the Docker stack
	cd $(INFRA) && docker compose down

.PHONY: infra-test
infra-test: ## Run the Docker integration smoke test
	cd $(INFRA) && ./scripts/integration-test.sh
