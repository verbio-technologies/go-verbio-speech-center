#!/bin/bash

set -euxo pipefail

apt update
apt install -y protobuf-compiler protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

$HOME/go/bin/protoc-gen-go-grpc
protoc-gen-go-grpc

