.PHONY: proto build run clean

proto:
	protoc --proto_path=api \
	       --go-grpc_out=paths=source_relative:gen/go \
	       api/login_service/auth.proto
	protoc --proto_path=api \
	       --go_out=paths=source_relative:gen/go \
	       api/login_service/auth.proto

build:
	go build -o build/login_service ./LoginService/cmd/login_service

run:
	go run ./LoginService/cmd/login_service/main.go

clean:
	rm -rf build/
	go clean
