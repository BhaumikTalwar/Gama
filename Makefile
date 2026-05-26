-include .env
export

GO ?= go
GOOS ?= linux
GOARCH ?= amd64
GOAMD64 ?= v3
CGO_ENABLED ?= 0

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

BIN_DIR = bin
MAIN_PKG = ./.
APP_NAME = gama
MODULE = github.com/BhaumikTalwar/Gama

DB_URL ?= "postgres://devuser:devpass@localhost:5432/devdb?sslmode=disable"
MIGRATION_DIR = internal/db/migrations

PROD_GIN_TAGS = jsoniter,nomsgpack

PROD_LDFLAGS = -s -w -buildid= \
  -X $(MODULE)/internal/buildinfo.Version=$(VERSION) \
  -X $(MODULE)/internal/buildinfo.Commit=$(COMMIT) \
  -X $(MODULE)/internal/buildinfo.BuildTime=$(BUILD_TIME)

PROD_GO_ENV = \
	CGO_ENABLED=$(CGO_ENABLED) \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	GOAMD64=$(GOAMD64)

PROD_GO_FLAGS = \
	-trimpath \
	-tags netgo,osusergo,$(PROD_GIN_TAGS) \
	-buildmode=pie \
	-ldflags "$(PROD_LDFLAGS)"


.PHONY: all init build prod run run-prod clean                    \
	migrate-up migrate-down migrate-status                        \                      
	docker-build docker-up docker-down docker-logs migrate-create \
	sqlc test-coverage docker-test-up
all: build

init:
	go mod tidy

sqlc:
	@echo "Generating sqlc code..."
	cd internal/db && sqlc generate

test-coverage:
	@echo "Running unit tests with cumulative coverage..."
	go test -cover -short -count=1 ./...
	@echo ""
	@echo "--- Generating cumulative profile ---"
	@mkdir -p coverage
	go test -coverprofile=coverage/unit.out -covermode=atomic -short -count=1 ./... > /dev/null 2>&1
	@echo "Total: $$(go tool cover -func=coverage/unit.out | grep '^total:' | awk '{print $$NF}')"
	@rm -f coverage/unit.out

test-coverage-html:
	@echo "Generating HTML coverage report..."
	@mkdir -p coverage
	go test -coverprofile=coverage/unit.out -covermode=atomic -short -count=1 ./... > /dev/null 2>&1
	go tool cover -html=coverage/unit.out -o coverage/coverage.html
	@echo "Report: coverage/coverage.html"
	@rm -f coverage/unit.out

test-coverage-all: docker-test-up
	@echo "Running all tests with cumulative coverage (requires Docker)..."
	go test -p=1 -tags=integration -cover -count=1 ./...
	@echo ""
	@echo "--- Generating cumulative profile ---"
	@mkdir -p coverage
	go test -p=1 -tags=integration -coverprofile=coverage/all.out -covermode=atomic -count=1 ./... > /dev/null 2>&1
	@echo "Total: $$(go tool cover -func=coverage/all.out | grep '^total:' | awk '{print $$NF}')"
	@rm -f coverage/all.out

test-unit:
	@echo "Running unit tests..."
	go test -short -count=1 ./...

test-integration: docker-test-up
	@echo "Starting test services..."
	docker-compose up -d postgres redis
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running integration tests (serial, shared database)..."
	go test -p=1 -tags=integration -count=1 -v ./internal/repository/ ./internal/auth/
	@echo "Stopping test services..."
	docker-compose stop postgres redis

test-all: docker-test-up
	@echo "Running all tests..."
	go test -p=1 -tags=integration -count=1 -v ./...

build: 
	$(GO)  mod tidy
	$(GO) build -v -x \
		-o $(BIN_DIR)/$(APP_NAME) \
		$(MAIN_PKG)

run: build
	redis-cli FLUSHALL
	./$(BIN_DIR)/$(APP_NAME)

prod: 
	$(GO) mod tidy
	$(PROD_GO_ENV) $(GO) build \
		$(PROD_GO_FLAGS) \
		-o $(BIN_DIR)/$(APP_NAME) \
		$(MAIN_PKG)

run-prod: prod
	./$(BIN_DIR)/$(APP_NAME)

clean:
	redis-cli FLUSHALL
	rm -f $(BIN_DIR)/$(APP_NAME)

migrate-up:
	goose -dir $(MIGRATION_DIR) postgres $(DB_URL) up

migrate-down:
	goose -dir $(MIGRATION_DIR) postgres $(DB_URL) down

migrate-status:
	goose -dir $(MIGRATION_DIR) postgres $(DB_URL) status

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir $(MIGRATION_DIR) create $$name sql

docker-test-up:
	docker-compose up -d postgres minio redis tempo prometheus

docker-build:
	docker-compose build

docker-build-no-cache:
	docker-compose build --no-cache

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

