VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build build-all clean install

build:
	go build -ldflags "$(LDFLAGS)" -o bin/orb cmd/orb/main.go

build-all:
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/orb-linux-amd64 cmd/orb/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/orb-linux-arm64 cmd/orb/main.go
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/orb-darwin-amd64 cmd/orb/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/orb-darwin-arm64 cmd/orb/main.go
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/orb-windows-amd64.exe cmd/orb/main.go

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/orb

clean:
	rm -rf bin/ dist/