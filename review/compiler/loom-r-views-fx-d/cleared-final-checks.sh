#!/usr/bin/env bash
set -u
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
ev=review/compiler/loom-r-views-fx-d
run_check() {
 local label=$1
 shift
 "$@" > "$ev/cleared-final-$label.log" 2>&1
 local result=$?
 printf '%s exit=%s\n' "$label" "$result"
 if [ "$result" -ne 0 ]; then return "$result"; fi
}
run_check focused timeout 240 go test ./internal/javascript ./internal/lower ./internal/native ./internal/oracle -run '^(TestFX7|TestLoomViewsD|TestCallableContract|TestViewCheckedCastMetadata$|TestRuntimeSetsAreEachUsed$)' -count=1 -v -timeout 55s || exit $?
run_check reader-guard timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s || exit $?
run_check vet timeout 150 go vet ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle || exit $?
run_check cli-build timeout 150 go build -o /tmp/loom-r-views-d ./cmd/adamic || exit $?
run_check replay timeout 360 python3 "$ev/replay.py" || exit $?
