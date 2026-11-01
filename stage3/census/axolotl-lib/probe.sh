#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
base=/workspace/scratch/axolotl
mkdir -p "$base/fixtures"
printf 'export {};\n' > "$base/fixtures/empty.ts"
printf 'import * as fs from "node:fs"; export const present = fs.existsSync(".");\n' > "$base/fixtures/node.ts"
git -C /workspace/adamic show HEAD:internal/load/node_library_fs.go > "$base/original_node_library_fs.go"
for mode in ${PROBE_MODES:-main es2020 es2020-capped-node}; do
 source=$base/compiler
 [[ $mode != main ]] || source=/workspace/adamic
 python3 - "$base" "$mode" "$source" <<'PY'
import json,pathlib,sys
base,mode,source=map(pathlib.Path,sys.argv[1:])
p=base/mode/'latent-overlay/overlay.json'
s=json.loads(p.read_text())
s['Replace'][str(source/'stage3/census/latent/tool/main.go')]=str(base/'compiler/stage3/census/axolotl_probe/main.go')
if str(mode)=='es2020':s['Replace'][str(source/'internal/load/node_library_fs.go')]=str(base/'original_node_library_fs.go')
(base/mode/'probe-overlay.json').write_text(json.dumps(s,indent=2)+'\n')
PY
 (cd "$source" && go build -buildvcs=false -overlay="$base/$mode/probe-overlay.json" -o "$base/$mode/probe" ./stage3/census/latent/tool) > "$base/$mode/probe-build.log" 2>&1
 for label in tsc empty node; do
  path=$base/fixtures/$label.ts
  [[ $label != tsc ]] || path=$base/adapted/src/tsc/tsc.ts
  "$base/$mode/probe" "$path" > "$base/$mode/probe-$label.jsonl" 2> "$base/$mode/probe-$label.log"
 done
done
