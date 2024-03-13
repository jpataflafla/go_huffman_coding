#Makefile to use for local testing (so password is inside repo)
#newgrp docker
#docker run --name go-huff-coding -e POSTGRES_PASSWORD="et@&d8XA2*%7nm" -p 5432:5432 -d postgres

# Build target: compiles the Go program and generates the binary in the bin directory
build:
	@go build -o bin/go_huffman_coding

# Run target: depends on the build target, runs the compiled binary
run: build
	@./bin/go_huffman_coding

# Test target: runs tests for the Go program with verbose output
test:
	@go test -v ./...
