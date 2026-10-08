#!/usr/bin/env bash
# A complete Node proof plus an optional native implementation of the same wire contract.
set -euo pipefail
scout=$(cd "$(dirname "$0")" && pwd)
if [[ $# -lt 2 || $# -gt 3 ]]; then
    echo 'usage: run.sh PINNED_ADAPTED_TREE NEW_RESULTS [NATIVE_BINDER_EXECUTABLE]' >&2
    exit 2
fi
: "${STEP31_TYPESCRIPT:?set STEP31_TYPESCRIPT to stock 6.0.3 lib/typescript.js}"
tree=$(realpath "$1")
results=$(realpath -m "$2")
[[ ! -e $results ]] || { echo "results must be new: $results" >&2; exit 2; }
mkdir -p "$results"
export PYTHONDONTWRITEBYTECODE=1
node "$scout/test.cjs" > "$results/fixtures.log" 2>&1
node "$scout/piece-probes.cjs" > "$results/probes.stdout" 2> "$results/probes.stderr"
python3 "$scout/compare.py" --output "$results/source-probes" "$results/probes.stdout" -- \
    node "$scout/source.mjs" "$tree" --probes > "$results/probe-comparison.log" 2>&1
python3 "$scout/probe-mutants.py" "$results/probe-mutants" > "$results/probe-mutants.log" 2>&1
python3 "$scout/check-compare.py" "$results/comparison-mutants" > "$results/comparison-mutants.log" 2>&1
python3 "$scout/mutants.py" "$tree" "$results/source-mutants" > "$results/source-mutants.log" 2>&1
python3 "$scout/prepare-corpus.py" "$results/projects.json" --upstream "$tree" > "$results/corpus.log" 2>&1
node "$scout/binder-dump.cjs" "$results/projects.json" > "$results/stock.stdout" 2> "$results/stock.stderr"
[[ ! -s $results/stock.stderr ]] || { echo 'stock stderr is not empty' >&2; exit 1; }
python3 "$scout/compare.py" --output "$results/source" "$results/stock.stdout" -- \
    node "$scout/source.mjs" "$tree" "$results/projects.json" > "$results/source-comparison.log" 2>&1
if [[ $# == 3 ]]; then
    native=$(realpath "$3")
    python3 "$scout/compare.py" --output "$results/native" "$results/stock.stdout" -- \
        "$native" "$results/projects.json" > "$results/native-comparison.log" 2>&1
fi
python3 - "$results" <<'PYTHON'
import hashlib, json, pathlib, sys
root = pathlib.Path(sys.argv[1])
request = json.loads((root / 'projects.json').read_text())
report = {'projects': len(request['projects']), 'acceptance': 301, 'upstream': len(request['projects']) - 301,
    'bytes': (root / 'stock.stdout').stat().st_size,
    'sha256': hashlib.sha256((root / 'stock.stdout').read_bytes()).hexdigest(),
    'node_source_matches_stock': json.loads((root / 'source/report.json').read_text())['success'],
    'native_run': (root / 'native/report.json').exists()}
(root / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report))
PYTHON
