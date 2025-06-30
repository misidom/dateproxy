.PHONY: build run test clean docker-build docker-run k8s-deploy k8s-clean

# Build the main proxy
build:
	go build -o bin/dateproxy .

# Build test services
build-services:
	go build -o bin/service1 ./test-backends/service1
	go build -o bin/service2 ./test-backends/service2
	go build -o bin/service3 ./test-backends/service3

# Run the proxy locally
run: build
	./bin/dateproxy -config config.yaml

# Run test services locally
run-service1:
	PORT=8081 go run ./test-backends/service1/main.go

run-service2:
	PORT=8082 go run ./test-backends/service2/main.go

run-service3:
	PORT=8083 go run ./test-backends/service3/main.go

# Test the proxy
test:
	@echo "Testing with date 20210515 (should route to service1)..."
	@curl -s "http://localhost:8080/api/test?date=20210515" | jq .
	@echo "\nTesting with date 20230815 (should route to service2)..."
	@curl -s "http://localhost:8080/api/test?date=20230815" | jq .
	@echo "\nTesting with date 20250315 (should route to service3)..."
	@curl -s "http://localhost:8080/api/test?date=20250315" | jq .

# Clean build artifacts
clean:
	rm -rf bin/

# Docker operations
docker-build:
	docker build -t dateproxy:latest .
	docker build -t service1:latest -f test-backends/service1/Dockerfile test-backends/service1/
	docker build -t service2:latest -f test-backends/service2/Dockerfile test-backends/service2/
	docker build -t service3:latest -f test-backends/service3/Dockerfile test-backends/service3/

# Kubernetes operations
k8s-deploy:
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/backend-services.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml

k8s-clean:
	kubectl delete -f k8s/ --ignore-not-found=true

# Development setup
dev-setup: build build-services
	@echo "Built all services. Run the following in separate terminals:"
	@echo "make run-service1"
	@echo "make run-service2"
	@echo "make run-service3"
	@echo "make run"