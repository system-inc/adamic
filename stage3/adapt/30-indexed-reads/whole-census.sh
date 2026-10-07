#!/usr/bin/env bash
set -euo pipefail
unit=$(cd "$(dirname "$0")" && pwd)
repo=$(git -C "$unit" rev-parse --show-toplevel)
tree=$(cd "$1" && pwd)
mkdir -p "$2"
out=$(cd "$2" && pwd)
python3 - "$repo" "$unit" "$out" <<'PY'
import json, re, subprocess, sys
from pathlib import Path
repo, unit, out = map(Path, sys.argv[1:])
pin = '176a496'
feature = subprocess.check_output(['git', '-C', str(repo), 'show', pin + ':stage3/census/latent/data/meter3/loader-options.go.txt'], text=True)
original = (repo / 'internal/load/load.go').read_text()
pattern = r'func compilerOptions\(\) \*core.CompilerOptions \{.*?\n\}'
before = re.search(pattern, original, re.S).group()
after = re.search(pattern, feature, re.S).group()
# Match run 0's archived options exactly, including its proven syntax features.
# No diagnostic code is filtered, and strict indexed/optional checking stays on.
expected = before.replace('ErasableSyntaxOnly:         core.TSTrue', 'ErasableSyntaxOnly:         core.TSFalse')
expected = expected.replace('\t\tNoImplicitReturns:          core.TSTrue,\n', '')
expected = expected.replace('\t\tNoFallthroughCasesInSwitch: core.TSTrue,\n', '')
assert expected == after, 'loader options drift from the pinned meter'
(out / 'meter-load.go').write_text(original.replace(before, after))
(out / 'overlay.json').write_text(json.dumps({'Replace': {
    str(repo / 'internal/load/load.go'): str(out / 'meter-load.go'),
    str(repo / 'cmd/adamic-meter/indexed_reads_export_test.go'): str(unit / 'census-export.go.txt'),
}}))
PY
cd "$repo"
ADAMIC_INDEXED_ROOT="$tree/src/compiler" ADAMIC_INDEXED_OUT="$out/all-diagnostics.json" \
    go test -overlay="$out/overlay.json" ./cmd/adamic-meter -run '^TestIndexedReadsCensus$' -count=1 -v > "$out/export.log" 2>&1
python3 - "$out" "$3" <<'PY' > "$out/counts.log" 2>&1
import json, re, sys
from pathlib import Path
out, file = Path(sys.argv[1]), sys.argv[2]
raw = json.loads((out / 'all-diagnostics.json').read_text())
pattern = r'.*/src/compiler/(' + re.escape(file) + r'):(\d+):(\d+): error TS(\d+):'
diagnostics = []
for text in raw['diagnostics']:
    match = re.match(pattern, text)
    if match:
        diagnostics.append('src/compiler/' + text[match.start(1):])
result = {'file': file, 'roots': raw['roots'], 'code_filter': None,
          'total': len(diagnostics), 'diagnostics': diagnostics}
(out / 'census.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2))
PY
