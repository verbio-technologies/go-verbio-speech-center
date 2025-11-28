#!/bin/bash

set -euxo pipefail

SCRIPT_PATH=$(readlink -f "$0")
REPO_ROOT=$(dirname "$SCRIPT_PATH")"/.."
PATH=${PATH}:${HOME}/go/bin

PROTO_DIR="${REPO_ROOT}/proto"
PROTO_FILENAMES="recognition_streaming_request.proto recognition_streaming_response.proto recognition.proto"

pushd "$REPO_ROOT"

for proto_file in ${PROTO_FILENAMES}; do

protoc \
      --proto_path="${PROTO_DIR}" \
      --go_out="${PROTO_DIR}" \
      --go-grpc_out="${PROTO_DIR}" \
      --experimental_allow_proto3_optional \
      "${PROTO_DIR}/${proto_file}"
done

popd