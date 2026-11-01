#!/usr/bin/env bash
set -euo pipefail
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
base=/workspace/scratch/axolotl
build() {
 local source=$1 mode=$2
 local out=$base/$mode
 mkdir -p "$out"
 (cd "$source" && go build -o "$out/census" ./stage3/census/tool) > "$out/build.log" 2>&1
 python3 /workspace/adamic/stage3/census/latent/make_overlay.py "$source" "$out/latent-overlay" > "$out/latent-overlay.log" 2>&1
 gofmt -w "$out/latent-overlay/"*.go
 python3 /workspace/adamic/stage3/meter/entry_overlay.py "$out/latent-overlay" "$out/entry-overlay" > "$out/entry-overlay.log" 2>&1
 gofmt -w "$out/entry-overlay/"*.go
 (cd "$source" && go build -buildvcs=false -overlay="$out/entry-overlay/overlay.json" -o "$out/entry-census" ./stage3/census/latent/tool) > "$out/entry-build.log" 2>&1
 (cd "$source" && go build -buildvcs=false -overlay="$out/latent-overlay/overlay.json" -o "$out/latent-census" ./stage3/census/latent/tool) > "$out/latent-build.log" 2>&1
 sha256sum "$out/census" "$out/entry-census" "$out/latent-census" > "$out/binaries.sha256"
}
measure() {
 local mode=$1
 local out=$base/$mode
 mkdir -p "$out/tsc"
 "$out/census" "$base/adapted/src/tsc/tsc.ts" "$out/tsc/census.jsonl" > "$out/tsc/census.log" 2>&1
 LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$out/entry-census" "$base/adapted/src/tsc/tsc.ts" "$out/tsc/latent.jsonl" > "$out/tsc/latent.log" 2>&1
 "$out/census" "$base/adapted/src/compiler" "$out/census.jsonl" > "$out/census.log" 2>&1
 LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$out/latent-census" "$base/adapted/src/compiler" "$out/latent.jsonl" > "$out/latent.log" 2>&1
}
build /workspace/adamic main
build "$base/compiler" es2020
measure main
measure es2020
python3 - <<'PY'
from pathlib import Path
p=Path('/workspace/scratch/axolotl/compiler/internal/load/node_library_fs.go')
s=p.read_text()
a='text, err := CorrectNodeHostAxioms(file, string(data))'
b=a+'\n\t\ttext = strings.ReplaceAll(text, `/// <reference lib="esnext.disposable" />`, "")\n\t\ttext = strings.ReplaceAll(text, `/// <reference lib="esnext.float16" />`, "")'
assert s.count(a)==1
p.write_text(s.replace(a,b))
PY
git -C "$base/compiler" diff -- internal/load > "$base/es2020-capped-node.patch"
build "$base/compiler" es2020-capped-node
measure es2020-capped-node
python3 /workspace/adamic/stage3/meter/box.py "$base/box.json" end
