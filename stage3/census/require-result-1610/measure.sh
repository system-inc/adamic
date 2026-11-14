#!/usr/bin/env bash
set -euo pipefail
s=/workspace/scratch/require-result-1610
export PATH=/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:$PATH
export TMPDIR=$s/tmp GOTMPDIR=$s/tmp GOCACHE=$s/gocache NODE_PATH=$s/work/stage3/api/node_modules GOMEMLIMIT=3GiB
while pgrep -f '^python3 /workspace/scratch/require-result-1610/work/stage3/apply.py ' >/dev/null; do sleep 5; done
test -f "$s/main-adapted/patch-set.md"
cp "$s/main-adapted/src/compiler/sys.ts" "$s/baseline-sys.ts"
node "$s/probe.cjs" > "$s/stock.log" 2>&1
cp "$s/main-adapted/src/compiler/sys.ts" "$s/candidate-sys.ts"
cp "$s/baseline-sys.ts" "$s/main-adapted/src/compiler/sys.ts"
node -e 'const fs=require("fs"),s=process.argv[1]; const a=JSON.parse(fs.readFileSync(s+"/stock-before.json")),b=JSON.parse(fs.readFileSync(s+"/stock-after.json")); if(b.diagnostics.length>a.diagnostics.length)process.exit(1)' "$s"
while pgrep -f '^go build .*require-result-1610/census' >/dev/null; do sleep 5; done
test -x "$s/census"
export LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=$s/main-adapted/src/compiler/sys.ts
"$s/census" "$s/main-adapted/src/compiler" "$s/main.jsonl" > "$s/census-main.log" 2>&1
cp "$s/candidate-sys.ts" "$s/main-adapted/src/compiler/sys.ts"
"$s/census" "$s/main-adapted/src/compiler" "$s/change.jsonl" > "$s/census-change.log" 2>&1
cp "$s/baseline-sys.ts" "$s/main-adapted/src/compiler/sys.ts"
"$s/census" "$s/main-adapted/src/compiler" "$s/restoring-mutant.jsonl" > "$s/census-mutant.log" 2>&1
printf 'measurements complete\n'
