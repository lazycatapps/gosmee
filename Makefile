# Example Makefile for a LazyCAT Apps project
# Copy this file to your project and customize the variables below

# Project configuration
# PROJECT_NAME ?= your-project  # defaults to current directory name
# Project type (lpk-only | docker-lpk)
PROJECT_TYPE ?= docker-lpk
APP_ID_PREFIX ?= cloud.lazycat.app.liu.

# Version (optional, auto-detected from git if not set)
# VERSION := 1.0.0

# Docker configuration (only for docker-lpk projects)
# REGISTRY := docker.io/lazycatapps
# IMAGE_NAME := $(PROJECT_NAME)

BACKEND_IMAGE ?= $(REGISTRY)/gosmee/backend
GO_MODULE_DIR ?= backend

# Include the common base.mk
include base.mk

# You can add custom targets below
# Example:
# .PHONY: custom-target
# custom-target: ## My custom target
#	@echo "Running custom target"

CLEAN_EXTRA_PATHS += dist backend/gosmee-web-server

DEV_DATA_DIR ?= /tmp/gosmee-data
DEV_MAX_CLIENTS_PER_USER ?= 50
DEV_MAX_STORAGE_PER_USER ?= 10737418240

.PHONY: dev-backend
dev-backend: ##@Development Run backend locally for development
	mkdir -p $(DEV_DATA_DIR) && \
	cd backend && \
	GOSMEE_DATA_DIR="$(DEV_DATA_DIR)" \
	GOSMEE_MAX_CLIENTS_PER_USER="$(DEV_MAX_CLIENTS_PER_USER)" \
	GOSMEE_MAX_STORAGE_PER_USER="$(DEV_MAX_STORAGE_PER_USER)" \
	go run cmd/server/main.go

.PHONY: dev-frontend
dev-frontend: ##@Development Run frontend locally for development
	cd frontend && npm start

.PHONY: audit
audit: ##@Development Scan frontend dependencies for vulnerabilities
	@echo "Scanning frontend dependencies for vulnerabilities..."
	@CURRENT_REGISTRY=$$(npm config get registry); \
	npm config set registry https://registry.npmjs.org/; \
	cd frontend && npm audit; \
	AUDIT_EXIT=$$?; \
	npm config set registry $$CURRENT_REGISTRY; \
	exit $$AUDIT_EXIT

.PHONY: push-backend
push-backend: ##@Maintenance Push production backend image to registry
	@$(MAKE) docker-push-default \
		DOCKER_BUILD_CONTEXT=backend \
		DOCKER_BUILD_TARGET=prod \
		DOCKER_BUILD_PLATFORM=$(PLATFORM) \
		FULL_IMAGE_NAME=$(BACKEND_IMAGE):$(VERSION)

.PHONY: push-backend-dev
push-backend-dev: build-local-bin ##@Maintenance Push development backend image to registry
	@$(MAKE) docker-push-default \
		DOCKER_BUILD_CONTEXT=backend \
		DOCKER_BUILD_TARGET=dev \
		DOCKER_BUILD_PLATFORM=$(PLATFORM) \
		FULL_IMAGE_NAME=$(BACKEND_IMAGE):$(VERSION)

.PHONY: build-frontend
build-frontend: ##@Build Build frontend into dist directory
	@echo "Building frontend to dist directory..."
	sh build.sh
	@echo "Frontend dist built successfully!"

.PHONY: build-local-bin
build-local-bin: ##@Build Build backend binary locally
	cd backend && CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o gosmee-web-server cmd/server/main.go

.PHONY: deploy
deploy: push-backend build-frontend deploy-default ##@Deploy Production deployment (backend prod + frontend + lpk)

.PHONY: deploy-frontend
deploy-frontend: build-frontend deploy-default ##@Deploy Deploy frontend only (frontend + lpk)

.PHONY: deploy-backend-dev
deploy-backend-dev: push-backend-dev deploy-default ##@Deploy Deploy backend dev only (backend dev + lpk)

.PHONY: deploy-full-dev
deploy-full-dev: push-backend-dev build-frontend deploy-default ##@Deploy Development full deployment (backend dev + frontend + lpk)
