test-server:
	@go test -v ./...

run-server:
	@go run ./cmd/main.go

build-server:
	@go build -o ./bin/server ./cmd/main.go