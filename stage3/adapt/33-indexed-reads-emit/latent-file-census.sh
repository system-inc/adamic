#!/usr/bin/env bash
# Use a scratch checker at the latent report's pinned taste branch.
set -euo pipefail
unit=$(cd "$(dirname "$0")" && pwd)
checker=$(cd "$1" && pwd)
tree=$(cd "$2" && pwd)
mkdir -p "$3"
out=$(cd "$3" && pwd)
python3 - "$unit" "$checker" "$out" <<'PY'
import json
from pathlib import Path
import subprocess
import sys
unit, checker, out = map(Path, sys.argv[1:])
# The report's other native feature merges do not change loader options.
# Compose its pinned fallthrough/implicit-return delta on its pinned taste loader.
source = subprocess.check_output(['git', 'show', 'd2c05df34443d9c0ae4fe6639beef2c5aee8a4ad:internal/load/load.go'], cwd=checker, text=True)
assert (checker / 'internal/load/load.go').read_text() == source
for line in ['\t\tNoImplicitReturns:          core.TSTrue,\n', '\t\tNoFallthroughCasesInSwitch: core.TSTrue,\n']:
    assert source.count(line) == 1
    source = source.replace(line, '')
# The removal is precisely 5f77d331's loader delta. Keep taste's syntax support.
assert '\t\tErasableSyntaxOnly:         core.TSFalse,' in source
(out / 'load.go').write_text(source)
export = (unit / 'census-export.go.txt').read_text()
assert export.count('package main') == 1
(out / 'export.go').write_text(export.replace('package main', 'package load_test', 1))
(out / 'overlay.json').write_text(json.dumps({'Replace': {
    str(checker / 'internal/load/load.go'): str(out / 'load.go'),
    str(checker / 'internal/load/indexed_reads_export_test.go'): str(out / 'export.go'),
}}))
PY
cd "$checker"
ADAMIC_INDEXED_ROOT="$tree/src/compiler" ADAMIC_INDEXED_OUT="$out/all-diagnostics.json" \
    go test -overlay="$out/overlay.json" ./internal/load -run '^TestIndexedReadsCensus$' -count=1 -v > "$out/export.log" 2>&1
