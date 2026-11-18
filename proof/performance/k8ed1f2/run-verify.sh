#!/usr/bin/env bash
# #k8ed1f2 roads 1 and 2: stock tsc over tsc's whole program with the edit list applied to main's apply tree, every
# diagnostic the baseline lacks with its source. Run after environment.sh, sealed-job.sh, sealed_typescript_source
# and sealed_npm stage3/api. Never merged.
set -u
A="$TMPDIR/adapted"
bash stage3/apply.sh "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }
VERIFY_KEEP_SERVICES=1 node proof/performance/k8ed1f2/verify.mjs "$A" proof/performance/k8ed1f2/edits-roads12.diff "$TMPDIR/verify" | grep -vE '^  kept '
