#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
out=stage3/census/map-object-keys-63979123
# The apply wrapper owns the live process. Wait for its completed resource record.
while [ ! -f "$out/apply.time.json" ]; do sleep 5; done
python3 - <<'PY'
import json
assert json.load(open('stage3/census/map-object-keys-63979123/apply.time.json'))['exit']==0
PY
source /workspace/adamic-tools/env.sh
node "$out/source-names.cjs" "$out/source-names.json" > "$out/source-names.log" 2>&1
bash "$out/checkpoint.sh" apply
