#!/usr/bin/env bash
# Scratch for #38kdd4d's W2 cause, never lands: the per-field table of writes through Mutable views, on apply's tree.
set -o pipefail
root=$(pwd)
STAGE3_CACHE="$TMPDIR/c" bash stage3/apply.sh "$TMPDIR/t" > "$TMPDIR/apply.log" 2>&1 || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
NODE_PATH="$root/stage3/api/node_modules" node scratch-w2/w2scout.cjs "$TMPDIR/t" scratch-w2/w2-rows-85.tsv "$TMPDIR/w2.json" || { echo "FAIL scout"; exit 1; }
echo "== json"
cat "$TMPDIR/w2.json"
