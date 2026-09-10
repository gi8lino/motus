# Makefile

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
DEV_PORT := $(LOCALBIN)/dev-port
OPEN_BROWSER := $(LOCALBIN)/open-browser
DEV_TAG := $(LOCALBIN)/dev-tag
GO_INSTALL_TOOL := $(LOCALBIN)/go-install-tool

# renovate: datasource=github-releases depName=gi8lino/dev-tools
DEV_TOOLS_VERSION ?= v0.3.0
DEV_PORT_VERSIONED := $(DEV_PORT)-$(DEV_TOOLS_VERSION)
OPEN_BROWSER_VERSIONED := $(OPEN_BROWSER)-$(DEV_TOOLS_VERSION)
DEV_TAG_VERSIONED := $(DEV_TAG)-$(DEV_TOOLS_VERSION)
GO_INSTALL_TOOL_VERSIONED := $(GO_INSTALL_TOOL)-$(DEV_TOOLS_VERSION)

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.13.2

# Default: no prefix. Can be overridden via `make patch VERSION_PREFIX=v`
VERSION_PREFIX ?= v

##@ Tagging

.PHONY: patch
patch: dev-tools ## Create a new patch release (x.y.Z+1).
	$(DEV_TAG) --prefix "$(VERSION_PREFIX)" patch

.PHONY: minor
minor: dev-tools ## Create a new minor release (x.Y+1.0).
	$(DEV_TAG) --prefix "$(VERSION_PREFIX)" minor

.PHONY: major
major: dev-tools ## Create a new major release (X+1.0.0).
	$(DEV_TAG) --prefix "$(VERSION_PREFIX)" major

.PHONY: tag
tag: dev-tools ## Show the latest tag.
	@echo "Latest version: $$($(DEV_TAG) --prefix "$(VERSION_PREFIX)" current)"

.PHONY: push
push: ## Push tags to the configured remote.
	git push --tags

##@ Development

# Persistent local ports, shared by separate Make invocations.
dev-port = $(or $(shell $(DEV_PORT) $(1)),$(error Could not resolve port for $(1)))
MOTUS_ASSIGNED_PORT ?= $(call dev-port,app)
DB_ASSIGNED_PORT ?= $(call dev-port,postgres)
SITE_ROOT ?= http://127.0.0.1:$(MOTUS_ASSIGNED_PORT)
RUN_ARGS ?=
COMPOSE_PROJECT ?= $(notdir $(CURDIR))
COMPOSE_FILE ?= deploy/motus/docker-compose.db.yml
COMPOSE = MOTUS_POSTGRES_PORT=$(DB_ASSIGNED_PORT) docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT)

.PHONY: ports ports-reset postgres serve dev-build
ports: dev-tools ## Print saved local development ports.
	@$(DEV_PORT) app --port "$(MOTUS_ASSIGNED_PORT)" > /dev/null
	@$(DEV_PORT) postgres --port "$(DB_ASSIGNED_PORT)" > /dev/null
	@echo "Motus: $(SITE_ROOT)/"
	@echo "Postgres: 127.0.0.1:$(DB_ASSIGNED_PORT)"

ports-reset: dev-tools ## Clear saved ports after stopping local services.
	$(DEV_PORT) --reset

postgres: ports ## Start local Postgres and wait for readiness.
	@$(COMPOSE) up -d --wait --wait-timeout 60 motus-db

dev-build: ports
	$(MAKE) web

serve: ports ## Run the app using the saved ports (build and Postgres must be ready).
	go run ./cmd --listen-address="127.0.0.1:$(MOTUS_ASSIGNED_PORT)" \
		--site-root="$(SITE_ROOT)" \
		--database-url="postgres://motus:motus@127.0.0.1:$(DB_ASSIGNED_PORT)/motus?sslmode=disable" \
		$(RUN_ARGS)


.PHONY: download
download: dev-tools ## Download go packages
	go mod download

.PHONY: run
run: dev-build postgres ## Build, start Postgres, and run Motus with the browser.
	@$(OPEN_BROWSER) "http://127.0.0.1:$(MOTUS_ASSIGNED_PORT)/" & \
	browser_pid=$$!; \
	trap 'kill "$$browser_pid" 2>/dev/null || true' EXIT; \
	$(MAKE) serve

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test test-backend test-frontend
test: test-backend test-frontend ## Run backend and frontend unit tests.

test-backend: vet ## Run backend unit tests.
	go test -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...

test-frontend: ## Run frontend unit tests.
	npm --prefix web test

.PHONY: cover
cover: ## Display test coverage
	go test -coverprofile=coverage.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Clean up generated files
	rm -f coverage.out coverage.html

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	$(GOLANGCI_LINT) run --fix

##@ Icons

SVG       := web/motus.svg
ICON_DIR  := web/public
ICON_SIZES := 16x16 32x32 48x48 64x64

.PHONY: favicon
favicon: ## Create favicons
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


##@ Admin

.PHONY: create-user
create-user: dev-tools ## Create a user via API: make create-user NAME="Alice"
	@[ -n "$(NAME)" ] || (echo "NAME is required" && exit 1)
	curl -sSf -X POST \
	  -H "Content-Type: application/json" \
	  -d '{"name":"$(NAME)"}' \
	  $(SITE_ROOT)/api/users

.PHONY: promote-admin
promote-admin: dev-tools ## Promote a user to admin via API: make promote-admin USER_ID=<id> ADMIN_USER_ID=<id>
	@[ -n "$(USER_ID)" ] || (echo "USER_ID is required" && exit 1)
	@[ -n "$(ADMIN_USER_ID)" ] || (echo "ADMIN_USER_ID is required (an existing admin)" && exit 1)
	curl -sSf -X PUT \
	  -H "Content-Type: application/json" \
	  -H "X-User-ID: $(ADMIN_USER_ID)" \
	  -d '{"isAdmin":true}' \
	  $(SITE_ROOT)/api/users/$(USER_ID)/admin

##@ Dependencies

.PHONY: dev-tools
dev-tools: \
	$(DEV_PORT_VERSIONED) \
	$(OPEN_BROWSER_VERSIONED) \
	$(DEV_TAG_VERSIONED) \
	$(GO_INSTALL_TOOL_VERSIONED) ## Download the pinned development tools.
	@ln -sf "$(notdir $(DEV_PORT_VERSIONED))" "$(DEV_PORT)"
	@ln -sf "$(notdir $(OPEN_BROWSER_VERSIONED))" "$(OPEN_BROWSER)"
	@ln -sf "$(notdir $(DEV_TAG_VERSIONED))" "$(DEV_TAG)"
	@ln -sf "$(notdir $(GO_INSTALL_TOOL_VERSIONED))" "$(GO_INSTALL_TOOL)"

$(DEV_PORT_VERSIONED): | $(LOCALBIN)
	$(call download-dev-tool,dev-port,$@)

$(OPEN_BROWSER_VERSIONED): | $(LOCALBIN)
	$(call download-dev-tool,open-browser,$@)

$(DEV_TAG_VERSIONED): | $(LOCALBIN)
	$(call download-dev-tool,dev-tag,$@)

$(GO_INSTALL_TOOL_VERSIONED): | $(LOCALBIN)
	$(call download-dev-tool,go-install-tool,$@)

# download-dev-tool downloads a versioned tool from gi8lino/dev-tools.
# $1 - release asset name
# $2 - versioned destination path
define download-dev-tool
	@set -eu; \
	tmp="$(2).tmp"; \
	trap 'rm -f "$$tmp"' EXIT INT TERM; \
	echo "Downloading gi8lino/dev-tools $(DEV_TOOLS_VERSION) $(1)"; \
	curl --fail --silent --show-error --location \
		"https://github.com/gi8lino/dev-tools/releases/download/$(DEV_TOOLS_VERSION)/$(1)" \
		-o "$$tmp"; \
	chmod +x "$$tmp"; \
	mv "$$tmp" "$(2)"; \
	trap - EXIT INT TERM
endef

.PHONY: golangci-lint
golangci-lint: dev-tools ## Download golangci-lint locally if necessary.
	$(GO_INSTALL_TOOL) \
		--target "$(GOLANGCI_LINT)" \
		--package github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
		--tool-version "$(GOLANGCI_LINT_VERSION)"

##@ Frontend

.PHONY: lint-web
lint-web: ## Run frontend eslint checks.
	npm --prefix web run lint

.PHONY: lint-fix-web
lint-fix-web: ## Run frontend eslint with auto-fix.
	npm --prefix web run lint:fix

.PHONY: web-install
web-install: ## Install frontend dependencies (prefers lockfile; falls back to npm install if out-of-sync).
	@npm --prefix web ci || ( \
		echo "npm ci failed (lockfile out-of-sync); running npm install to refresh web/package-lock.json"; \
		npm --prefix web install \
	)

.PHONY: web-ci
web-ci: ## Strict frontend install from lockfile (for CI).
	npm --prefix web ci

.PHONY: web
web: ## Build frontend app.
	npm --prefix web run build


##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: open
open: ports ## Open the browser once the application responds.
	$(OPEN_BROWSER) "http://127.0.0.1:$(MOTUS_ASSIGNED_PORT)/"
