.PHONY: build build-all build-employee build-business build-gateway run-employee run-business run-gateway docker-up docker-down docker-logs fmt tidy test kill clean

build-all: build-employee build-business build-gateway

build-employee:
	cd backend && go build -o ../build/employee_service ./cmd/employee

build-business:
	cd backend && go build -o ../build/business_service ./cmd/business

build-gateway:
	cd backend && go build -o ../build/gateway ./cmd/gateway

run-employee:
	cd backend && go run ./cmd/employee/main.go

run-business:
	cd backend && go run ./cmd/business/main.go

run-gateway:
	cd backend && go run ./cmd/gateway/main.go

test:
	cd backend && go test -v -race ./...

docker-up:
	cd backend && docker compose up -d

docker-down:
	cd backend && docker compose down

docker-logs:
	cd backend && docker compose logs -f

fmt:
	cd backend && go fmt ./...

tidy:
	cd backend && go mod tidy

kill:
	lsof -i :8080 -i :8081 -i :8082 | grep LISTEN | awk '{print $$2}' | xargs kill -9 2>/dev/null || true

clean:
	rm -rf build/
	go clean
