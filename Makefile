.PHONY: build test clean

BINARY_NAME=githelper
VERSION=$(shell git describe --tags --always --dirty)
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildDate=${BUILD_DATE}"

build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME}

test:
	go test -v ./...

clean:
	rm -rf bin/ 