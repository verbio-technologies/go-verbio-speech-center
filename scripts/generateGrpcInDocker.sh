#!/bin/bash

set -euxo pipefail

SCRIPT_PATH=$(readlink -f $0)
REPO_ROOT=$(echo "$SCRIPT_PATH" | sed 's|scripts/.*||g')

IMAGE="golang:1.17.5-buster"
WD="/code"
CURRENT_UID=$(id -u $(whoami))
CURRENT_GID=$(id -g $(whoami))

pushd $REPO_ROOT
docker run -it --rm \
  -v "${REPO_ROOT}":"${WD}" \
  --workdir="${WD}" \
  -e EXTERNAL_UID=$CURRENT_UID \
  -e EXTERNAL_GID=$CURRENT_GID \
  $IMAGE \
  scripts/dockerGrpc.sh
popd