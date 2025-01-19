run-server:
	@CGO_ENABLED=1 go run ./cmd/main.go

build-server:
	@CGO_ENABLED=1 go build -o ./bin/server ./cmd/main.go