#!/usr/bin/env bash
# #r11wppy item 1's build candidate type-checked. Run after environment.sh, sealed-job.sh, sealed_typescript_source
# and sealed_npm stage3/api. Never merged.
set -u
A="$TMPDIR/adapted"
bash stage3/apply.sh "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }
node proof/performance/r11wppy-fixpoint/verify.mjs "$A" proof/performance/r11wppy-fixpoint/edits3.diff "$TMPDIR/verify"
