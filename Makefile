.PHONY: proto build build-all run clean kill stop fmt

# ============= Build =============
build-all:
	cd backend && go build -o ../build/identity_service ./cmd/identity && \
	go build -o ../build/business_service ./cmd/business && \
	go build -o ../build/logs_service ./cmd/logs && \
	go build -o ../build/gateway ./cmd/gateway

build-identity:
	cd backend && go build -o ../build/identity_service ./cmd/identity

build-business:
	cd backend && go build -o ../build/business_service ./cmd/business

build-logs:
	cd backend && go build -o ../build/logs_service ./cmd/logs

build-gateway:
	cd backend && go build -o ../build/gateway ./cmd/gateway

# ============= Proto =============
proto-identity:
	protoc --proto_path=backend/api \
		--go_out=paths=source_relative:backend/gen/go \
		--go-grpc_out=paths=source_relative:backend/gen/go \
		backend/api/identity/identity.proto

proto-business:
	protoc --proto_path=backend/api \
		--go_out=paths=source_relative:backend/gen/go \
		--go-grpc_out=paths=source_relative:backend/gen/go \
		backend/api/business/business.proto

proto: proto-identity proto-business

# ============= Run =============
run-identity:
	cd backend && go run ./cmd/identity/main.go

run-business:
	cd backend && go run ./cmd/business/...

run-logs:
	cd backend && go run ./cmd/logs/main.go

run-gateway:
	cd backend && go run ./cmd/gateway/main.go

# ============= Docker =============
docker-up:
	cd backend && docker-compose up -d

docker-down:
	cd backend && docker-compose down

docker-logs:
	cd backend && docker-compose logs -f

# ============= Utils =============
fmt:
	cd backend && go fmt ./...

tidy:
	cd backend && go mod tidy

kill:
	lsof -i :50051 -i :50052 -i :8080 | grep LISTEN | awk '{print $$2}' | xargs kill -9 2>/dev/null || true

clean:
	rm -rf build/
	go clean
