# ===== App =====
APP_NAME := bot

.PHONY: run test lint
run:
	go run ./cmd/$(APP_NAME)

test:
	go test ./...

lint:
	golangci-lint run ./...

# ===== DB =====
DB_DSN ?= postgres://crnbot:crnbot@localhost:5432/crnbot?sslmode=disable
MIGRATIONS_DIR := cmd/$(APP_NAME)/migrations

# ===== Tools =====
# golang-migrate CLI. Must be built with -tags 'postgres' for postgres support.
MIGRATE_BIN := $(shell go env GOPATH)/bin/migrate

.PHONY: tools
tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ===== Docker =====
.PHONY: db-up db-down db-logs db-ps
db-up:
	docker compose up -d db

db-down:
	docker compose down

db-logs:
	docker compose logs -f db

db-ps:
	docker compose ps

# ===== Migrations =====
.PHONY: migrate-up migrate-down migrate-drop migrate-force migrate-version migrate-create

migrate-up: tools
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

# Rollback 1 migration
migrate-down: tools
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down 1

# DANGEROUS: drops everything in the database used by DB_DSN
migrate-drop: tools
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" drop -f

# Use only if DB is in "dirty" state. Example: make migrate-force V=1
migrate-force: tools
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" force $(V)

migrate-version: tools
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" version

# Create new migration files:
# make migrate-create NAME=add_users
migrate-create: tools
	@test -n "$(NAME)" || (echo "NAME is required: make migrate-create NAME=..." && exit 1)
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATIONS_DIR) $(NAME)
