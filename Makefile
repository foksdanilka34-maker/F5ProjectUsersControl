.PHONY: proto build build-all run clean kill-port stop
 
build-all:
	cd backend && go build -o ../build/login_service ./LoginService/cmd/login_service && \
	go build -o ../build/employee_service ./EmployeeService/cmd/employee_service && \
	go build -o ../build/project_service ./ProjectService/cmd/project_service && \
	go build -o ../build/analytics_service ./AnalyticsService/cmd/analytics_service && \
	go build -o ../build/api_gateway ./ApiGateway/cmd/api-gateway

proto-login:
	protoc --proto_path=backend/api \
	    --go-grpc_out=paths=source_relative:backend/gen/go \
	    backend/api/login_service/auth.proto
	protoc --proto_path=backend/api \
	    --go_out=paths=source_relative:backend/gen/go \
	    backend/api/login_service/auth.proto

proto-empl:
	protoc --proto_path=backend/api \
	    --go-grpc_out=paths=source_relative:backend/gen/go \
	    backend/api/employee_service/employee.proto
	protoc --proto_path=backend/api \
	    --go_out=paths=source_relative:backend/gen/go \
	    backend/api/employee_service/employee.proto

proto-proj:
	protoc --proto_path=backend/api \
	    --go-grpc_out=paths=source_relative:backend/gen/go \
	    backend/api/project_service/project.proto
	protoc --proto_path=backend/api \
	    --go_out=paths=source_relative:backend/gen/go \
	    backend/api/project_service/project.proto

proto-analytics:
	protoc --proto_path=backend/api \
	    --go-grpc_out=paths=source_relative:backend/gen/go \
	    backend/api/analytics_service/analytics.proto
	protoc --proto_path=backend/api \
	    --go_out=paths=source_relative:backend/gen/go \
	    backend/api/analytics_service/analytics.proto

fmt:
	cd backend && go fmt ./...

build:
	cd backend && go build -o ../build/login_service ./LoginService/cmd/login_service

run-ls:
	cd backend && go run ./LoginService/cmd/login_service/main.go
run-es:
	cd backend && go run ./EmployeeService/cmd/employee_service/main.go
run-ps:
	cd backend && go run ./ProjectService/cmd/project_service/main.go
run-as:
	cd backend && go run ./AnalyticsService/cmd/analytics_service/main.go
run-ag:
	cd backend && go run ./ApiGateway/cmd/api-gateway/main.go

kill:
	lsof -i :50051 | grep LISTEN | awk '{print $$2}' | xargs kill -9 2>/dev/null || true

stop:
	pkill -f "login_service" || true

clean:
	rm -rf build/
	go clean
