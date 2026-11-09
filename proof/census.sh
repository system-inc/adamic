#!/usr/bin/env bash
# The census tests that failed on Codex in 0.01s, at main fd8b74b9 (this commit's base) and here, to tell the environment from the change.
set -u
source /workspace/adamic-tools/env.sh 2>/dev/null || true
mkdir -p /tmp/adamic-gate
here=$(pwd)
git worktree add --detach /tmp/basewt fd8b74b9bd > /dev/null 2>&1 && (cd /tmp/basewt && rmdir cohere 2>/dev/null; ln -sfn $here/cohere cohere)
for d in $here /tmp/basewt; do echo "=== $d"; (cd $d && go test -count=1 -timeout 10m -v -run 'TestCycleFinderCensusEquality$|TestProgramRegionCensusMembership$' ./internal/lower/ 2>&1 | grep -E -- "--- (PASS|FAIL)|_test.go:[0-9]+:" | head -6 | cut -c1-300); done
