# Makefile

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin

## Tool Versions
# renovate: datasource=github-releases depName=gi8lino/dev-tools
DEV_TOOLS_VERSION ?= v0.7.0

## Tool Binaries
DEV_TOOL_NAMES := dev-port open-browser dev-tag make-help go-install-tool
DEV_TOOL_TARGETS := $(addprefix $(LOCALBIN)/,$(DEV_TOOL_NAMES))
DEV_TOOL_VERSIONED := $(addsuffix -$(DEV_TOOLS_VERSION),$(DEV_TOOL_TARGETS))

DEV_PORT := $(LOCALBIN)/dev-port
OPEN_BROWSER := $(LOCALBIN)/open-browser
DEV_TAG := $(LOCALBIN)/dev-tag
MAKE_HELP := $(LOCALBIN)/make-help
GO_INSTALL_TOOL := $(LOCALBIN)/go-install-tool

# Run a local tool while displaying only its executable name.
define run-tool
@printf '%s\n' '$(notdir $(1)) $(2)'
@$(1) $(2)
endef

$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries


GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.13.2

##@ Tagging

VERSION_PREFIX ?= v

.PHONY: current
current: $(DEV_TAG) ## Show the current semantic version tag.
	$(call run-tool,$(DEV_TAG),--prefix "$(VERSION_PREFIX)" current)

.PHONY: patch
patch: $(DEV_TAG) ## Create a new patch release (x.y.Z+1).
	$(call run-tool,$(DEV_TAG),--prefix "$(VERSION_PREFIX)" patch)

.PHONY: minor
minor: $(DEV_TAG) ## Create a new minor release (x.Y+1.0).
	$(call run-tool,$(DEV_TAG),--prefix "$(VERSION_PREFIX)" minor)

.PHONY: major
major: $(DEV_TAG) ## Create a new major release (X+1.0.0).
	$(call run-tool,$(DEV_TAG),--prefix "$(VERSION_PREFIX)" major)

.PHONY: tag
tag: current

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
ports: $(DEV_PORT) ## Print saved local development ports.
	@$(DEV_PORT) app --port "$(MOTUS_ASSIGNED_PORT)" > /dev/null
	@$(DEV_PORT) postgres --port "$(DB_ASSIGNED_PORT)" > /dev/null
	@echo "Motus: $(SITE_ROOT)/"
	@echo "Postgres: 127.0.0.1:$(DB_ASSIGNED_PORT)"

ports-reset: $(DEV_PORT) ## Clear saved ports after stopping local services.
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
help: $(MAKE_HELP) ## Display this help.
	@$(MAKE_HELP) $(MAKEFILE_LIST)

.PHONY: open
open: ports ## Open the browser once the application responds.
	$(OPEN_BROWSER) "http://127.0.0.1:$(MOTUS_ASSIGNED_PORT)/"

##@ Development tools

.PHONY: dev-tools
dev-tools: $(DEV_TOOL_TARGETS) ## Download the pinned development tools.

$(DEV_TOOL_TARGETS): $(LOCALBIN)/%: $(LOCALBIN)/%-$(DEV_TOOLS_VERSION)
	@ln -sf "$(notdir $<)" "$@"

$(DEV_TOOL_VERSIONED): $(LOCALBIN)/%-$(DEV_TOOLS_VERSION): | $(LOCALBIN)
	$(call download-dev-tool,$*,$@)

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
golangci-lint: $(GO_INSTALL_TOOL) ## Download golangci-lint locally if necessary.
	@$(GO_INSTALL_TOOL) \
		--target "$(GOLANGCI_LINT)" \
		--package github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
		--tool-version "$(GOLANGCI_LINT_VERSION)"
