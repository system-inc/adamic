#!/usr/bin/env bash
# Focused component proof; stdout, stderr, exits and timings are recorded separately.
set -euo pipefail
scout=$(cd "$(dirname "$0")" && pwd)
if [[ $# -lt 2 || $# -gt 4 ]]; then
    echo 'usage: run-components.sh PINNED_ADAPTED_TREE NEW_RESULTS [NATIVE_CHECKER [NATIVE_EMITTER]]' >&2
    exit 2
fi
: "${STEP31_TYPESCRIPT:?set STEP31_TYPESCRIPT to stock 6.0.3 lib/typescript.js}"
tree=$(realpath "$1")
results=$(realpath -m "$2")
[[ ! -e $results ]] || { echo "results must be new: $results" >&2; exit 2; }
mkdir -p "$results"
export PYTHONDONTWRITEBYTECODE=1
node "$scout/test-components.cjs" > "$results/fixtures.log" 2>&1
python3 "$scout/test-component-cache.py" "$results/cache-tests" > "$results/cache-tests.log" 2>&1
python3 "$scout/component-mutants.py" "$tree" "$results/source-mutants" > "$results/source-mutants.log" 2>&1
python3 "$scout/check-compare.py" "$results/process-mutants" > "$results/process-mutants.log" 2>&1
python3 "$scout/prepare-corpus.py" "$results/projects.json" --upstream "$tree" > "$results/corpus.log" 2>&1
python3 "$scout/benchmark-components.py" "$results/projects.json" "$results/benchmark" > "$results/benchmark.log" 2>&1
for mode in checker emitter; do
    python3 "$scout/compare.py" --timeout 180 --output "$results/source-$mode" \
        "$results/benchmark/$mode-full/golden.stdout" -- \
        node "$scout/source.mjs" "$tree" "--$mode" "$results/projects.json" > "$results/source-$mode.log" 2>&1
done
if [[ $# -ge 3 ]]; then
    python3 "$scout/component-compare.py" "$results/benchmark/checker-full" "$results/native-checker" -- \
        "$(realpath "$3")" > "$results/native-checker.log" 2>&1
fi
if [[ $# == 4 ]]; then
    python3 "$scout/component-compare.py" "$results/benchmark/emitter-full" "$results/native-emitter" -- \
        "$(realpath "$4")" > "$results/native-emitter.log" 2>&1
fi
python3 - "$results" <<'PYTHON'
import json, pathlib, sys
root = pathlib.Path(sys.argv[1])
report = {'projects': 6563, 'acceptance': 301, 'upstream': 6262,
    'source_matches_stock': {mode: json.loads((root / ('source-' + mode) / 'report.json').read_text())['success'] for mode in ['checker', 'emitter']},
    'native_run': {mode: (root / ('native-' + mode) / 'report.json').exists() for mode in ['checker', 'emitter']},
    'benchmark': json.loads((root / 'benchmark/report.json').read_text())}
(root / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps({key: value for key, value in report.items() if key != 'benchmark'}))
PYTHON
