.PHONY: all build dev run clean frontend backend image

# Default target
all: build

# Build everything
build: frontend backend

# Build frontend
frontend:
	cd frontend && pnpm install && pnpm run build

# Build backend (includes embedded frontend)
backend:
	go build -o supafirehose .

# Run in development mode (frontend hot reload + backend)
dev:
	@echo "Starting development servers..."
	@echo "Run 'cd frontend && npm run dev' in another terminal"
	go run . --dev

# Run production build
run: build
	./supafirehose

# Clean build artifacts
clean:
	rm -f supafirehose
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# Initialize database
db-init:
	psql -h localhost -U postgres -d pooler_demo -f init.sql

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...
	cd frontend && pnpm run format 2>/dev/null || true

# Tidy dependencies
tidy:
	go mod tidy
	cd frontend && pnpm install

# Build Docker image
image: build
	docker build -t supafirehose:latest .
