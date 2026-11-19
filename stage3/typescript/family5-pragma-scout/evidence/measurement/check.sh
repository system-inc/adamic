#!/usr/bin/env bash
set -euo pipefail
d=/workspace/scratch/family5-measure
api=$d/main/stage3/api/node_modules
tsc=$api/typescript/lib/tsc.js
e=$d/report/stage3/typescript/family5-pragma-scout/evidence/measurement
mkdir -p "$e"
node "$tsc" --version > "$e/tsc-version.txt"
: > "$e/tsc-exit-status.txt"
for label in before after; do
    cd "$d/$label"
    test -L node_modules || ln -s "$api" node_modules
    node "$tsc" -p src/compiler --showConfig --pretty false > "$d/$label-stock-config.json"
    python3 - "$d/$label-stock-config.json" <<'PY'
import json, pathlib, sys
p = json.load(open(sys.argv[1]))
roots = ['src/compiler/' + name.removeprefix('./') for name in p['files']]
expected = sorted(str(f) for f in pathlib.Path('src/compiler').rglob('*.ts'))
assert sorted(roots) == expected and len(roots) == 79, (len(roots), len(expected), set(roots)^set(expected))
pathlib.Path('loader-tsconfig.json').write_text(json.dumps({'compilerOptions': {
    'strict': True, 'noUncheckedIndexedAccess': True, 'exactOptionalPropertyTypes': True,
    'erasableSyntaxOnly': False, 'verbatimModuleSyntax': True, 'allowImportingTsExtensions': True,
    'noEmit': True, 'module': 'esnext', 'moduleDetection': 'force',
    'moduleResolution': 'bundler', 'target': 'es2024', 'lib': ['es2024'], 'types': [],
    'pretty': False, 'noErrorTruncation': True,
}, 'files': roots}, indent=2) + '\n')
PY
    status=0
    node "$tsc" -p src/compiler --pretty false --noErrorTruncation --tsBuildInfoFile "$d/$label-stock.tsbuildinfo" > "$e/$label-stock.txt" 2>&1 || status=$?
    printf '%s stock %s\n' "$label" "$status" >> "$e/tsc-exit-status.txt"
    status=0
    node "$tsc" -p loader-tsconfig.json > "$e/$label-strict.txt" 2>&1 || status=$?
    printf '%s strict %s\n' "$label" "$status" >> "$e/tsc-exit-status.txt"
done
cp "$d/after/loader-tsconfig.json" "$e/loader-tsconfig.json"
cp "$d/after-stock-config.json" "$e/stock-config.json"
