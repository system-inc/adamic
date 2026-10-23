#!/bin/bash
# run.sh <worktree> <label> <go test args...>: one cold run, fresh empty build cache, log to logs/<label>.log
set -u
WT="$1"; LABEL="$2"; shift 2
S=/tmp/claude-0/-home-user-adamic/f8a9298f-6bb4-53b5-9b8d-8a8c2435c1d1/scratchpad
export PATH=/opt/llvm20/bin:/opt/node24/bin:/usr/local/go/bin:/usr/bin:/bin
export TMPDIR=/tmp/adamic-gate
unset ADAMIC_BUILD_OUTPUT_DIR
CACHE=$(mktemp -d /tmp/adamic-gate/cache-XXXXXX); chmod 755 "$CACHE"
export ADAMIC_BUILD_CACHE_DIR="$CACHE"
LOG="$S/logs/$LABEL.log"
{ echo "# label $LABEL"; echo "# cwd $WT/internal/native"; echo "# ADAMIC_BUILD_CACHE_DIR=$CACHE (fresh, empty)"; echo "# loadavg before: $(cat /proc/loadavg)"; echo "# command: go test ./internal/native $*"; } > "$LOG"
cd "$WT" && START=$(date +%s.%N)
go test ./internal/native "$@" >> "$LOG" 2>&1; CODE=$?
END=$(date +%s.%N)
{ echo "# exit $CODE"; echo "# wall $(echo "$END - $START" | bc)s"; echo "# loadavg after: $(cat /proc/loadavg)"; } >> "$LOG"
rm -rf "${CACHE:?}"
grep -E '^(--- |ok|FAIL|# exit|# wall)|phase ' "$LOG"
