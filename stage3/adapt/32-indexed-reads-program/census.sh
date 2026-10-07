#!/usr/bin/env bash
set -euo pipefail
unit=$(cd "$(dirname "$0")" && pwd)
repo=${CENSUS_COMPILER_REPO:-$(git -C "$unit" rev-parse --show-toplevel)}
package=${CENSUS_PACKAGE:-cmd/adamic-meter}
tree=$(cd "$1" && pwd)
mkdir -p "$2"
out=$(cd "$2" && pwd)
python3 - "$repo" "$unit" "$out" "$package" <<'PY'
import json
from pathlib import Path
import sys
repo, unit, out = map(Path, sys.argv[1:4])
package = sys.argv[4]
(out / "overlay.json").write_text(json.dumps({"Replace": {
    str(repo / package / "indexed_reads_export_test.go"): str(unit / "census-export.go.txt")
}}))
PY
cd "$repo"
ADAMIC_INDEXED_ROOT="$tree/src/compiler" ADAMIC_INDEXED_OUT="$out/all-diagnostics.json" \
    go test -overlay="$out/overlay.json" "./$package" -run '^TestIndexedReadsCensus$' -count=1 -v > "$out/export.log" 2>&1
python3 "$unit/census.py" "$out/all-diagnostics.json" "$out/census.json" > "$out/counts.log" 2>&1
