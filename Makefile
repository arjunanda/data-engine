.PHONY: build build-all clean test

# Build for current platform
build:
	cd go && go build -ldflags "-s -w" -o ../bin/data-engine .

# Build for all platforms
build-all:
	mkdir -p bin
	cd go && GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ../bin/data-engine-linux-amd64 .
	cd go && GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o ../bin/data-engine-linux-arm64 .
	cd go && GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o ../bin/data-engine-darwin-amd64 .
	cd go && GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o ../bin/data-engine-darwin-arm64 .
	cd go && GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ../bin/data-engine-windows-amd64.exe .

# Run tests
test:
	cd go && go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/*

# Install dependencies
deps:
	cd go && go mod download
	npm install
