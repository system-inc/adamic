#!/usr/bin/env bash
set -euo pipefail
cd /tmp/polling-queue-roads
out=$PWD/stage3/census/polling-queue-roads/evidence
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/polling-typescript
# Prerequisites: the complete apply and census build in METHOD.md have succeeded.
node - <<'JS'
const fs=require('fs'),{roads}=require('./stage3/census/polling-queue-roads/roads.cjs');if(!fs.readFileSync('/tmp/polling-adapted/src/compiler/sys.ts','utf8').replace(/\r\n/g,'\n').includes(roads.original))throw Error('apply incomplete or pinned owner drift');
JS
test -x /tmp/polling-census
cp /tmp/polling-apply.log "$out/apply.log"
cp /tmp/polling-overlay.log "$out/overlay.log"
cp /tmp/polling-build.log "$out/build.log"
rg -n 'PollingIntervalQueue|\.pollScheduled\b|\.pollIndex\b|\.pollingInterval\b' /tmp/polling-adapted/src /tmp/polling-adapted/tests > "$out/adapted-search.txt"
node stage3/census/polling-queue-roads/witness.cjs /tmp/polling-adapted > "$out/witness.log"
for road in original a b c; do
 node stage3/census/polling-queue-roads/measure.cjs /tmp/polling-adapted "$road" > "$out/typecheck.$road.log" 2>&1
 LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=/tmp/polling-adapted/src/compiler/sys.ts /tmp/polling-census /tmp/polling-adapted/src/compiler "$out/census.$road.jsonl" > "$out/census.$road.log" 2>&1
 echo "$road completed"
done
cp "$out/sys.original.ts" /tmp/polling-adapted/src/compiler/sys.ts
python3 stage3/census/polling-queue-roads/compare.py
