.PHONY: help all build deps cross-compile cross-compile-docker fmt test coverage vet clean build_all

ARTIFACT_NAME:=speech_center

SHELL:=/bin/bash
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_ROOT := $(dir $(MAKEFILE_PATH))
BIN_DIRECTORY := ${PROJECT_ROOT}/bin

GO:=go

DISABLE_CACHE:=-count=1
VERBOSE:=-v
TEST_COMMAND:=${GO} test ./...
TEST_NO_CACHE_COMMAND:=${TEST_COMMAND} ${DISABLE_CACHE}
BUILD_COMMAND:=${GO} build
SRC_DIR:=${PROJECT_ROOT}

VERSION=`git describe --tags --always --dirty`
VERSION_VARIABLE_NAME=verbio_speech_center/constants.VERSION
VERSION_VARIABLE_BUILD_FLAG=-ldflags "-X ${VERSION_VARIABLE_NAME}=${VERSION}"
BUILD_WITH_VERSION_COMMAND=CGO_ENABLED=0 ${BUILD_COMMAND} ${VERSION_VARIABLE_BUILD_FLAG}

all: help

deps: ## Download all dependencies
	@ ${GO} get verbio_speech_center

build: deps speech_center ## Builds the binaries

speech_center: deps ## Builds the binary
	@ ${BUILD_WITH_VERSION_COMMAND} -o ${BIN_DIRECTORY}/speech_center cmd/speech_center/main.go

version: ## Print the version
	@echo $(VERSION)

test: ## Run unit tests
	@ ${GO} test ./... -v -count=1 # count=1 means disable test cache

grpc: ## Generate GRPC files
	@ scripts/generateGrpc.sh

grpc-docker: ## Generate GRPC files (using Docker, useful if you don't have the dependencies installed)
	@ scripts/generateGrpcInDocker.sh

coverage: ## Run tests with coverage
	@ scripts/coverage.sh

fmt: ## Apply linting and formatting
	@ ${GO} fmt ./...

check-fmt: ## Complain if any file does not comply to format guidelines
	@ test -z "$(gofmt -s -l $(find . -name '*.go' -type f -print) | tee /dev/stderr)"

vet: ## Run go vet
	@ ${GO} vet ./...

clean: ## Remove all build artifacts
	@ rm -rf ${BIN_DIRECTORY}

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
