.PHONY: help dev dev-logs dev-down clean test-auth

help:
	@echo "Health Bar - Docker Commands"
	@echo "make dev         - Start database and auth service"
	@echo "make dev-logs    - View logs"
	@echo "make dev-down    - Stop services"
	@echo "make clean       - Remove all containers and volumes"

dev:
	docker-compose -f docker-compose.dev.yml up -d

dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f

dev-down:
	docker-compose -f docker-compose.dev.yml down

clean:
	docker-compose -f docker-compose.dev.yml down -v
	docker system prune -f

test-auth:
	@echo "Testing Auth Service..."
	@sleep 2
	@curl -s -X POST http://localhost:8001/api/auth/register -H "Content-Type: application/json" -d '{"email":"patient1@test.com","password":"password123","role":"patient"}'
	@echo ""
	@curl -s -X POST http://localhost:8001/api/auth/login -H "Content-Type: application/json" -d '{"email":"patient1@test.com","password":"password123"}'
