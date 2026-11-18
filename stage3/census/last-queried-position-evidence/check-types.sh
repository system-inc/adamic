#!/usr/bin/env bash
set -eu
tree=$1
repo=$2
adamic=$3
wrapper=$4
logs=$5
label=$6
mapfile -t roots < <(find "$tree/src/compiler" -type f \( -name '*.ts' -o -name '*.a' \) | LC_ALL=C sort)
printf '%s\n' "${roots[@]}" > "$logs/$label-roots.txt"
[[ ${#roots[@]} == 79 ]]
ln -s "$repo/stage3/typescript/node_modules" "$tree/node_modules"
python3 - "$tree" "${roots[@]}" <<'PY'
import json, sys
from pathlib import Path
tree, *roots = sys.argv[1:]
Path(tree, 'proof-tsconfig.json').write_text(json.dumps({
    'extends': './src/compiler/tsconfig.json',
    'compilerOptions': {'strictFunctionTypes': True, 'noEmit': True, 'incremental': False, 'composite': False, 'emitDeclarationOnly': False, 'pretty': False},
    'files': roots,
    'include': [],
}, indent=2) + '\n')
PY
tsc="$repo/stage3/api/node_modules/typescript/lib/tsc.js"
[[ $(node "$tsc" --version) == 'Version 6.0.3' ]]
stock=0
bash "$wrapper" "$label stock tsc 6.0.3 (79 roots)" "$logs/$label-stock.log" node "$tsc" -p "$tree/proof-tsconfig.json" || stock=$?
types=0
bash "$wrapper" "$label adamic types (79 roots)" "$logs/$label-adamic.log" "$adamic" types "${roots[@]}" || types=$?
if (( stock != 0 )); then exit "$stock"; fi
exit "$types"
