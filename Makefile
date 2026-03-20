# ----------------------------------------------------------------------------
# DragnCards - Makefile
#
# DEFAULT STACK (compose.yml)
#   make run          - start the default compose stack
#   make stop         - stop the default compose stack (containers kept)
#   make clean        - remove stopped containers and project images
#   make add-user EMAIL=x@y.com PASSWORD=secret
#
# OFFLINE STACK (compose.offline.yml)
#   make offline-build    - build images and save .tar files to ./build/
#   make offline-install  - load .tar images and start the offline stack
#   make offline-run      - start the offline compose stack
#   make offline-stop     - stop the offline compose stack (containers kept)
#   make offline-clean    - remove build artefacts and project images
#   make offline-add-user EMAIL=x@y.com PASSWORD=secret [ALIAS=nick]
#
# SUB-PATH DEPLOYMENT
#   To serve the frontend at example.com/dragncards instead of the root:
#     1. In compose.offline.yml, uncomment and set: NGINX_BASE_PATH=/dragncards
#     2. Run:     make offline-run   (no rebuild needed – subpath is applied at startup)
# ----------------------------------------------------------------------------

BACKEND_IMAGE  := dragncards/backend:latest
FRONTEND_IMAGE := dragncards/frontend:latest
POSTGRES_IMAGE := postgres:16
NGINX_IMAGE    := nginx:latest

BUILD_DIR       := build
DEFAULT_COMPOSE := compose.yml
OFFLINE_COMPOSE := compose.offline.yml

.PHONY: all help \
        run stop clean add-user \
        offline-build offline-install offline-run offline-stop offline-clean offline-add-user \
        _ensure_build_dir

all: help

## -- HELP ---------------------------------------------------------------------
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
     /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 }' \
	     $(MAKEFILE_LIST)
	@echo ""

# =============================================================================
# DEFAULT STACK  (compose.yml)
# =============================================================================

## -- run ----------------------------------------------------------------------
run: ## Start the default compose stack
	docker compose -f $(DEFAULT_COMPOSE) up -d
	@echo "==> Default stack is up."
	@echo "    Frontend : http://localhost:3000"
	@echo "    Backend  : http://localhost:4000"

## -- stop ---------------------------------------------------------------------
stop: ## Stop the default compose stack (containers kept, data preserved)
	docker compose -f $(DEFAULT_COMPOSE) stop

## -- clean --------------------------------------------------------------------
clean: ## Remove stopped default containers and project Docker images
	@echo "==> Removing stopped containers ..."
	docker compose -f $(DEFAULT_COMPOSE) down --remove-orphans
	@echo "==> Removing project Docker images ..."
	docker rmi -f $(BACKEND_IMAGE) $(FRONTEND_IMAGE) 2>/dev/null || true
	@echo "==> Done."

## -- add-user -----------------------------------------------------------------
add-user: ## Create a user (default stack)  ->  make add-user EMAIL=x@y.com PASSWORD=secret [ALIAS=nick]
ifndef EMAIL
	$(error EMAIL is required. Usage: make add-user EMAIL=user@example.com PASSWORD=secret [ALIAS=nick])
endif
ifndef PASSWORD
	$(error PASSWORD is required. Usage: make add-user EMAIL=user@example.com PASSWORD=secret [ALIAS=nick])
endif
	@echo "$(EMAIL) $(PASSWORD) $(ALIAS)" | docker compose -f $(DEFAULT_COMPOSE) exec -T backend \
		mix run /app/priv/batch_create_users.exs

# =============================================================================
# OFFLINE STACK  (compose.offline.yml)
# =============================================================================

## -- offline-build ------------------------------------------------------------
offline-build: _ensure_build_dir ## Build images and save to ./build/ (run on online machine)
	@echo "==> Building backend image ..."
	docker build -t $(BACKEND_IMAGE) ./backend

	@echo "==> Building frontend image ..."
	docker build -t $(FRONTEND_IMAGE) ./frontend

	@echo "==> Pulling postgres image ..."
	docker pull $(POSTGRES_IMAGE)

	@echo "==> Pulling nginx image ..."
	docker pull $(NGINX_IMAGE)

	@echo "==> Saving images to $(BUILD_DIR)/ ..."
	docker save $(BACKEND_IMAGE)  -o $(BUILD_DIR)/backend.tar
	docker save $(FRONTEND_IMAGE) -o $(BUILD_DIR)/frontend.tar
	docker save $(POSTGRES_IMAGE) -o $(BUILD_DIR)/postgres.tar
	docker save $(NGINX_IMAGE)    -o $(BUILD_DIR)/nginx.tar

	@echo "==> Done. Transfer the ./$(BUILD_DIR)/ folder and this repo to the offline machine."

_ensure_build_dir:
	mkdir -p $(BUILD_DIR)

## -- offline-install ----------------------------------------------------------
offline-install: ## Load .tar images into Docker and start the offline stack
	@echo "==> Loading images into Docker ..."
	docker load -i $(BUILD_DIR)/backend.tar
	docker load -i $(BUILD_DIR)/frontend.tar
	docker load -i $(BUILD_DIR)/postgres.tar
	docker load -i $(BUILD_DIR)/nginx.tar

	@echo "==> Starting compose stack ($(OFFLINE_COMPOSE)) ..."
	docker compose -f $(OFFLINE_COMPOSE) up -d

	@echo "==> Offline stack is up."
	@echo "    Frontend : http://localhost:3000"
	@echo "    Backend  : http://localhost:4000"
	@echo "    Images   : http://localhost:8080"

## -- offline-run --------------------------------------------------------------
offline-run: ## Start the offline compose stack
	docker compose -f $(OFFLINE_COMPOSE) up -d
	@echo "==> Offline stack is up."
	@echo "    Frontend : http://localhost:3000"
	@echo "    Backend  : http://localhost:4000"
	@echo "    Images   : http://localhost:8080"

## -- offline-stop -------------------------------------------------------------
offline-stop: ## Stop the offline compose stack (containers kept, data preserved)
	docker compose -f $(OFFLINE_COMPOSE) stop

## -- offline-clean ------------------------------------------------------------
offline-clean: ## Remove build/*.tar artefacts and offline containers/images
	@echo "==> Removing offline containers ..."
	docker compose -f $(OFFLINE_COMPOSE) down --remove-orphans
	@echo "==> Removing build artefacts ..."
	rm -rf $(BUILD_DIR)
	@echo "==> Removing project Docker images ..."
	docker rmi -f $(BACKEND_IMAGE) $(FRONTEND_IMAGE) 2>/dev/null || true
	@echo "==> Done."

## -- offline-add-user ---------------------------------------------------------
offline-add-user: ## Create a user (offline stack)  ->  make offline-add-user EMAIL=x@y.com PASSWORD=secret [ALIAS=nick]
ifndef EMAIL
	$(error EMAIL is required. Usage: make offline-add-user EMAIL=user@example.com PASSWORD=secret [ALIAS=nick])
endif
ifndef PASSWORD
	$(error PASSWORD is required. Usage: make offline-add-user EMAIL=user@example.com PASSWORD=secret [ALIAS=nick])
endif
	@echo "$(EMAIL) $(PASSWORD) $(ALIAS)" | docker compose -f $(OFFLINE_COMPOSE) exec -T backend \
		mix run /app/priv/batch_create_users.exs

## -- ha-be lint -------------------------------------------------------------
ha-be-lint: ## Run golangci-lint in ha-be
	cd ha-be && golangci-lint run --config .golangci.yml ./...
