BINARY     := ./analyzer.elf
CONFIG     := config/dev.yaml
SERVER_ADDR := localhost:50051
PROJECT_ID := 1

.PHONY: demo build db-up db-down migrate-up migrate-down server-start server-stop test-grpc clean

# ─── Full demo ────────────────────────────────────────────────────────────────

demo: build db-up migrate-up server-start test-grpc server-stop migrate-down db-down

# ─── Build ────────────────────────────────────────────────────────────────────

build:
	@echo "==> Building binary..."
	go build -o $(BINARY) ./cmd

# ─── Database ─────────────────────────────────────────────────────────────────

db-up:
	@echo "==> Starting PostgreSQL..."
	docker compose up -d postgres
	@echo "==> Waiting for PostgreSQL to be ready..."
	@until docker compose exec postgres pg_isready -U postgres > /dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "==> PostgreSQL is ready."

db-down:
	@echo "==> Stopping PostgreSQL..."
	docker compose down

# ─── Migrations ───────────────────────────────────────────────────────────────

migrate-up:
	@echo "==> Applying all migrations..."
	$(BINARY) migrate --config $(CONFIG)

migrate-down:
	@echo "==> Rolling back all migrations..."
	$(BINARY) migrate --config $(CONFIG) --down 0

# ─── gRPC server ──────────────────────────────────────────────────────────────

server-start:
	@echo "==> Starting gRPC server in background..."
	$(BINARY) serve --config $(CONFIG) > /tmp/analyzer-server.log 2>&1 & echo $$! > /tmp/analyzer-server.pid
	@echo "==> Waiting for gRPC server to be ready..."
	@for i in $$(seq 1 20); do \
		if $(BINARY) client --addr $(SERVER_ADDR) get-stats --project-id $(PROJECT_ID) > /dev/null 2>&1; then \
			echo "==> gRPC server is ready."; \
			break; \
		fi; \
		sleep 1; \
		if [ $$i -eq 20 ]; then \
			echo "ERROR: gRPC server did not start in time"; \
			cat /tmp/analyzer-server.log; \
			exit 1; \
		fi; \
	done

server-stop:
	@echo "==> Stopping gRPC server..."
	@if [ -f /tmp/analyzer-server.pid ]; then \
		kill $$(cat /tmp/analyzer-server.pid) 2>/dev/null || true; \
		rm -f /tmp/analyzer-server.pid; \
	fi

# ─── gRPC client tests ────────────────────────────────────────────────────────

test-grpc:
	@echo ""
	@echo "══════════════════════════════════════════"
	@echo "  Testing gRPC endpoints"
	@echo "══════════════════════════════════════════"

	@echo ""
	@echo "── GetStats (project $(PROJECT_ID)) ──────────────────"
	$(BINARY) client --addr $(SERVER_ADDR) get-stats --project-id $(PROJECT_ID)

	@echo ""
	@echo "── GetChart: open_state_histogram ────────"
	$(BINARY) client --addr $(SERVER_ADDR) get-chart \
		--project-id $(PROJECT_ID) --chart-type open_state_histogram

	@echo ""
	@echo "── GetChart: state_distribution ──────────"
	$(BINARY) client --addr $(SERVER_ADDR) get-chart \
		--project-id $(PROJECT_ID) --chart-type state_distribution

	@echo ""
	@echo "── GetChart: complexity_histogram ────────"
	$(BINARY) client --addr $(SERVER_ADDR) get-chart \
		--project-id $(PROJECT_ID) --chart-type complexity_histogram

	@echo ""
	@echo "── GetChart: priority ────────────────────"
	$(BINARY) client --addr $(SERVER_ADDR) get-chart \
		--project-id $(PROJECT_ID) --chart-type priority

	@echo ""
	@echo "── GetChart: daily_activity ──────────────"
	$(BINARY) client --addr $(SERVER_ADDR) get-chart \
		--project-id $(PROJECT_ID) --chart-type daily_activity

	@echo ""
	@echo "── CompareProjects ($(PROJECT_ID) vs $(PROJECT_ID)) ───────────────"
	$(BINARY) client --addr $(SERVER_ADDR) compare \
		--project-id-a $(PROJECT_ID) --project-id-b $(PROJECT_ID)

	@echo ""
	@echo "══════════════════════════════════════════"
	@echo "  All endpoints tested successfully"
	@echo "══════════════════════════════════════════"

# ─── Cleanup ──────────────────────────────────────────────────────────────────

clean: server-stop
	@echo "==> Cleaning up..."
	rm -f $(BINARY)
	docker compose down -v
