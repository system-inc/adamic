#!/usr/bin/env bash
set -eu
source /workspace/adamic-tools/env.sh
git apply review/test-defend/internal-native-decode_ascii/D1.diff
trap 'git restore -- internal/native/devirtualize.go' EXIT
go vet ./internal/native/ > /tmp/defend-native-vet.log 2>&1
ADAMIC_BUILD_CACHE_DIR=/tmp/defend-native/cache/replay-D1 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run "$(cat review/test-defend/internal-native-decode_ascii/matrix.regex)" > /tmp/defend-native-D1.log 2>&1
