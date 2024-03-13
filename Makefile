# Build target: compiles the Go program and generates the binary in the bin directory
build:
	@go build -o bin/go_huffman_coding

# Run target: depends on the build target, runs the compiled binary
run: build
	@./bin/go_huffman_coding

# Test target: runs tests for the Go program with verbose output
test:
	@go test -v ./...
