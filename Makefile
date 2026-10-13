.DEFAULT_GOAL := help

## Tool Versions

# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.14.0

# renovate: datasource=github-releases depName=gi8lino/dev-tools
DEV_TOOLS_VERSION ?= v0.9.0

NPX ?= npx

## Shared development tools

include bin/dev-tools.mk
include $(call dev-tools-module,tag)
include $(call dev-tools-module,port)
include $(call dev-tools-module,browser)
include $(call dev-tools-module,help)

## Project-local tools

GOLANGCI_LINT := bin/golangci-lint

## Build Configuration

BINARY ?= motus
COMMAND ?= ./cmd/motus
BUILD_VERSION ?= dev
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS ?= -s -w -X main.Version=$(BUILD_VERSION) -X main.Commit=$(BUILD_COMMIT)
RUN_ARGS ?=

## Development

COMPOSE_PROJECT ?= $(notdir $(CURDIR))
COMPOSE_FILE := deploy/motus/docker-compose.db.yml
DB_CONTAINER_NAME ?= motus-db

MOTUS_ASSIGNED_PORT ?= $(call dev-port,app)
MOTUS_MANAGEMENT_ASSIGNED_PORT ?= $(call dev-port,management)
DB_ASSIGNED_PORT ?= $(call dev-port,postgres)
PUBLIC_URL ?= http://127.0.0.1:$(MOTUS_ASSIGNED_PORT)

## Frontend

NPM ?= npm
WEB_NODE_MODULES := web/node_modules/.package-lock.json

##@ Development

.PHONY: ports-reset
ports-reset: $(DEV_PORT) ## Clear saved ports after stopping local services.
	$(call run-tool,$(DEV_PORT),--reset)

.PHONY: ports
ports: $(DEV_PORT) ## Print selected local development ports.
	@$(DEV_PORT) app --port "$(MOTUS_ASSIGNED_PORT)" > /dev/null
	@$(DEV_PORT) management --port "$(MOTUS_MANAGEMENT_ASSIGNED_PORT)" > /dev/null
	@$(DEV_PORT) postgres --port "$(DB_ASSIGNED_PORT)" > /dev/null
	@echo "Motus: $(PUBLIC_URL)/"
	@echo "Management: http://127.0.0.1:$(MOTUS_MANAGEMENT_ASSIGNED_PORT)/"
	@echo "Postgres: 127.0.0.1:$(DB_ASSIGNED_PORT)"

.PHONY: postgres
postgres: ports ## Run PostgreSQL locally.
	@echo "Starting PostgreSQL on dynamic host port: $(DB_ASSIGNED_PORT)"
	@MOTUS_POSTGRES_PORT=$(DB_ASSIGNED_PORT) docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) up -d --wait --wait-timeout 60 $(DB_CONTAINER_NAME)

.PHONY: dev-build
dev-build: ports web ## Build development assets.

.PHONY: serve
serve: ports ## Run Motus using the saved ports.
	@echo "Starting Motus application..."
	@go run $(COMMAND) \
		--listen-address="127.0.0.1:$(MOTUS_ASSIGNED_PORT)" \
		--management-listen-address="127.0.0.1:$(MOTUS_MANAGEMENT_ASSIGNED_PORT)" \
		--public-url="$(PUBLIC_URL)" \
		--log-format=text \
		--database-url="postgres://motus:motus@127.0.0.1:$(DB_ASSIGNED_PORT)/motus?sslmode=disable" \
		$(RUN_ARGS)

.PHONY: open
open: ports $(OPEN_BROWSER) ## Open Motus in the browser.
	$(call run-tool,$(OPEN_BROWSER),"$(PUBLIC_URL)/")

.PHONY: run
run: dev-build postgres $(OPEN_BROWSER) ## Build, start services, and run Motus.
	@$(OPEN_BROWSER) "$(PUBLIC_URL)/" & \
	browser_pid=$$!; \
	trap 'kill "$$browser_pid" 2>/dev/null || true' EXIT; \
	$(MAKE) serve RUN_ARGS="$(RUN_ARGS)"

##@ Build

.PHONY: build
build: web ## Build the Motus binary.
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(COMMAND)

.PHONY: vet
vet: ## Run Go static analysis.
	go vet ./...

.PHONY: clean
clean: ## Clean generated application files.
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf web/dist

##@ Testing

.PHONY: test
test: test-backend test-frontend ## Run backend and frontend unit tests.

.PHONY: test-backend
test-backend: vet ## Run backend unit tests.
	go test -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...

.PHONY: test-frontend
test-frontend: $(WEB_NODE_MODULES) ## Run frontend unit tests.
	$(NPM) --prefix web test

.PHONY: cover
cover: ## Display Go test coverage.
	go test -coverprofile=coverage.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...
	go tool cover -html=coverage.out

##@ Assets

SVG := web/motus.svg
ICON_DIR := web/public
ICON_SIZES := 16x16 32x32 48x48 64x64

.PHONY: favicon
favicon: ## Create favicons.
	@mkdir -p $(ICON_DIR)
	@for size in $(ICON_SIZES); do \
		outfile=favicon-$$size.png; \
		echo "create $$outfile"; \
		convert $(SVG) \
			-fuzz 5% -transparent white \
			-background none \
			-resize $$size \
			$(ICON_DIR)/$$outfile \
			>/dev/null 2>&1; \
	done
	@echo "create apple-touch-icon.png"
	@convert $(SVG) \
		-fuzz 5% -transparent white \
		-background none \
		-resize 180x180 \
		$(ICON_DIR)/apple-touch-icon.png \
		>/dev/null 2>&1

##@ Frontend

.PHONY: web
web: $(WEB_NODE_MODULES) ## Build the frontend application.
	$(NPM) --prefix web run build

.PHONY: lint-web
lint-web: $(WEB_NODE_MODULES) ## Run frontend eslint checks.
	$(NPM) --prefix web run lint

.PHONY: lint-fix-web
lint-fix-web: $(WEB_NODE_MODULES) ## Run frontend eslint with auto-fix.
	$(NPM) --prefix web run lint:fix

##@ Formatting

.PHONY: fmt
fmt: fmt-go fmt-web ## Format authored source.

.PHONY: fmt-go
fmt-go: ## Format Go code.
	go fmt ./...

.PHONY: fmt-web
fmt-web: ## Format CSS and TypeScript source.
	$(NPX) prettier --write \
		"web/src/**/*.css" \
		"web/src/**/*.ts" \
		"web/src/**/*.tsx" \
		"web/tests/*.ts" \
		"web/tests/**/*.ts"


##@ Linting

.PHONY: lint
lint: lint-go lint-web ## Run all linters.

.PHONY: lint-go
lint-go: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint and apply fixes.
	$(GOLANGCI_LINT) run --fix

##@ Dependencies

$(WEB_NODE_MODULES): web/package.json web/package-lock.json
	$(NPM) --prefix web ci

.PHONY: download
download: dev-tools $(WEB_NODE_MODULES) ## Download project dependencies.
	go mod download

.PHONY: dev-tools
dev-tools: $(DEV_PORT) $(OPEN_BROWSER) $(DEV_TAG) $(MAKE_HELP) $(GO_INSTALL_TOOL) ## Download pinned development tools.

.PHONY: golangci-lint
golangci-lint: $(GO_INSTALL_TOOL) ## Download golangci-lint locally if necessary.
	@$(GO_INSTALL_TOOL) \
		--target "$(GOLANGCI_LINT)" \
		--package github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
		--tool-version "$(GOLANGCI_LINT_VERSION)"
