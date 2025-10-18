.PHONY: help run build test docker-build docker-run compose-up compose-down k8s-deploy k8s-delete

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally
	@echo "🚀 Starting application..."
	go run cmd/api/main.go

build: ## Build the application binary
	@echo "🔨 Building binary..."
	go build -o bin/api cmd/api/main.go
	@echo "✓ Binary created at bin/api"

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t pdf-ingestion:latest .
	@echo "✓ Docker image built"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 \
		-e POSTGRES_HOST=host.docker.internal \
		-e MONGO_HOST=host.docker.internal \
		-e MINIO_ENDPOINT=host.docker.internal:9000 \
		pdf-ingestion:latest

compose-up: ## Start all services with Docker Compose
	@echo "🐳 Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✓ Services started"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  MongoDB: localhost:27017"
	@echo "  MinIO: localhost:9000 (console: localhost:9001)"

compose-down: ## Stop all services
	@echo "🛑 Stopping services..."
	docker-compose down
	@echo "✓ Services stopped"

compose-logs: ## Show logs from Docker Compose services
	docker-compose logs -f

k8s-deploy: ## Deploy to Kubernetes
	@echo "☸️  Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/postgres/
	kubectl apply -f k8s/mongodb/
	kubectl apply -f k8s/minio/
	kubectl apply -f k8s/api/
	@echo "✓ Deployment complete"

k8s-delete: ## Delete Kubernetes resources
	@echo "🗑️  Deleting Kubernetes resources..."
	kubectl delete -f k8s/api/ --ignore-not-found
	kubectl delete -f k8s/minio/ --ignore-not-found
	kubectl delete -f k8s/mongodb/ --ignore-not-found
	kubectl delete -f k8s/postgres/ --ignore-not-found
	kubectl delete -f k8s/namespace.yaml --ignore-not-found
	@echo "✓ Resources deleted"

k8s-status: ## Check Kubernetes deployment status
	@echo "📊 Kubernetes Status:"
	kubectl get pods -n default
	kubectl get svc -n default

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf bin/
	@echo "✓ Clean complete"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies downloaded"

