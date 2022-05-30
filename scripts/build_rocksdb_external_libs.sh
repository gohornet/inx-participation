#!/bin/bash
#
# Builds the app with the latest commit hash (short)
# E.g.: ./app -v --> APPNAME 75316fe

DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

commit_hash=$(git rev-parse --short HEAD)
CGO_ENABLED=1 go build -ldflags="-s -w -X github.com/iotaledger/inx-participation/core/app.Version=$commit_hash" -tags rocksdb,external_libs
