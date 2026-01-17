.PHONY: help dev dev-logs dev-down clean test-auth db-shell db-tables db-view

help:
	@echo "Health Bar - Docker Commands"
	@echo "make dev         - Start database and auth service"
	@echo "make dev-logs    - View logs"
	@echo "make dev-down    - Stop services"
	@echo "make clean       - Remove all containers and volumes\nmake db-shell    - Enter PostgreSQL shell\nmake db-tables   - List all database tables\nmake db-view     - View table data (usage: make db-view table=users)"

dev:
	docker-compose -f docker-compose.dev.yml up -d

dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f

dev-down:
	docker-compose -f docker-compose.dev.yml down

clean:
	docker-compose -f docker-compose.dev.yml down -v
	docker system prune -f

db-shell:
	docker exec -it healthbar-postgres psql -U postgres -d healthbar

db-tables:
	docker exec healthbar-postgres psql -U postgres -d healthbar -c "\dt"

db-view:
	@if [ -z "$(table)" ]; then \
		echo "Error: 'table' argument is required."; \
		echo "Usage: make db-view table=<table_name>"; \
		echo "Example: make db-view table=users"; \
	else \
		docker exec healthbar-postgres psql -U postgres -d healthbar -c "SELECT * FROM $(table);"; \
	fi

test-auth:
	@echo "Testing Auth Service..."
	@sleep 2
	@curl -s -X POST http://localhost:8001/api/auth/register -H "Content-Type: application/json" -d '{"email":"patient1@test.com","password":"password123","role":"patient"}'
	@echo ""
	@curl -s -X POST http://localhost:8001/api/auth/login -H "Content-Type: application/json" -d '{"email":"patient1@test.com","password":"password123"}'
