#!/bin/bash

set -euxo pipefail

SCRIPT_PATH=$(readlink -f $0)
REPO_ROOT=$(echo "$SCRIPT_PATH" | sed 's|scripts/.*||g')

if [[ -z "$EXTERNAL_UID" ]]; then
  echo "EXTERNAL_UID variable must be set"
  exit 1
fi

if [[ -z "$EXTERNAL_GID" ]]; then
  echo "EXTERNAL_GID variable must be set"
  exit 1
fi

pushd $REPO_ROOT
scripts/installDependencies.sh
scripts/generateGrpc.sh
chown -R $EXTERNAL_UID:$EXTERNAL_GID proto/speech_center
popd