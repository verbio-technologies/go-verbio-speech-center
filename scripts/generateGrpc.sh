#!/bin/bash

set -euxo pipefail

SCRIPT_PATH=$(readlink -f $0)
REPO_ROOT=$(echo "$SCRIPT_PATH" | sed 's|scripts/.*||g')

PROTO_DIR="proto"
PROTO_FILENAME="grpc_gateway.proto"
MODULE_NAME="verbio-speech-center/proto"

mkdir -p "${PROTO_DIR}/speech_center"
pushd $REPO_ROOT
protoc \
    --go_opt=paths=source_relative \
    --go_opt="M${PROTO_FILENAME}=${MODULE_NAME}" \
    --proto_path="${PROTO_DIR}" \
    --go-grpc_out="${PROTO_DIR}" \
    --go_out="${PROTO_DIR}/speech_center" \
    "${PROTO_DIR}/${PROTO_FILENAME}"

popd
