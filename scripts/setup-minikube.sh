#!/bin/bash

echo "🚀 Setting up Minikube for PDF Ingestion Service"
echo "================================================"

# Check if minikube is installed
if ! command -v minikube &> /dev/null; then
    echo "❌ Minikube is not installed. Please install it first:"
    echo "   https://minikube.sigs.k8s.io/docs/start/"
    exit 1
fi

# Start minikube
echo "→ Starting Minikube..."
minikube start --cpus=4 --memory=8192 --driver=docker

# Enable addons
echo "→ Enabling addons..."
minikube addons enable metrics-server
minikube addons enable dashboard

# Build Docker image inside minikube
echo "→ Building Docker image inside Minikube..."
eval $(minikube docker-env)
docker build -t pdf-ingestion:latest .

# Deploy to Kubernetes
echo "→ Deploying to Kubernetes..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres/
kubectl apply -f k8s/mongodb/
kubectl apply -f k8s/minio/

# Wait for databases to be ready
echo "→ Waiting for databases to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n pdf-ingestion --timeout=300s
kubectl wait --for=condition=ready pod -l app=mongodb -n pdf-ingestion --timeout=300s
kubectl wait --for=condition=ready pod -l app=minio -n pdf-ingestion --timeout=300s

# Deploy API
echo "→ Deploying API service..."
kubectl apply -f k8s/api/

# Wait for API to be ready
echo "→ Waiting for API to be ready..."
kubectl wait --for=condition=ready pod -l app=pdf-api -n pdf-ingestion --timeout=300s

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Service Status:"
kubectl get pods -n pdf-ingestion
echo ""
kubectl get svc -n pdf-ingestion

echo ""
echo "🌐 To access the API:"
echo "   Run: minikube service pdf-api -n pdf-ingestion"
echo ""
echo "📊 To access the dashboard:"
echo "   Run: minikube dashboard"

