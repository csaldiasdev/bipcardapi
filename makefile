SERVER_MAIN_GOFILE_PATH = ./cmd/main.go

build-server:
	go build -o bin/server $(SERVER_MAIN_GOFILE_PATH)

run-server:
	go run $(SERVER_MAIN_GOFILE_PATH)