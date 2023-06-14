all:
	make build exec
build:
	go build -o bin/ ./cmd/chirpy

exec:
	./bin/chirpy
