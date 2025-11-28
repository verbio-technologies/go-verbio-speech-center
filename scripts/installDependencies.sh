#!/bin/bash

set -euxo pipefail

sudo apt update
sudo apt install -y protobuf-compiler protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


