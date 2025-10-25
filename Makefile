.PHONY: proto build run clean kill-port stop

proto-login:
	protoc --proto_path=api \
	       --go-grpc_out=paths=source_relative:gen/go \
	       api/login_service/auth.proto
	protoc --proto_path=api \
	       --go_out=paths=source_relative:gen/go \
	       api/login_service/auth.proto

proto-empl:
	protoc --proto_path=api \
	       --go-grpc_out=paths=source_relative:gen/go \
	       api/employee_service/employee.proto
	protoc --proto_path=api \
	       --go_out=paths=source_relative:gen/go \
	       api/employee_service/employee.proto

fmt:
	go fmt ./...

build:
	go build -o build/login_service ./LoginService/cmd/login_service

run-ls:
	go run ./LoginService/cmd/login_service/main.go
run-es:
	go run ./EmployeeService/cmd/employee_service/main.go

kill:
	lsof -i :50051 | grep LISTEN | awk '{print $$2}' | xargs kill -9 2>/dev/null || true

stop:
	pkill -f "login_service" || true

clean:
	rm -rf build/
	go clean
