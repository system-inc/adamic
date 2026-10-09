#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
if [ "$#" -ne 2 ]; then
    echo "usage: prove-mutant.sh <new output dir> <stock tsc.js>" >&2
    exit 2
fi
export PERFORMANCE_MUTANT_TSC=$(realpath "$2")
set +e
"$root/run.sh" "$1" --native "$root/diagnostic-mutant.py"
status=$?
set -e
python3 - "$1" "$status" <<'PY'
import json
from pathlib import Path
import sys
out = Path(sys.argv[1])
assert sys.argv[2] == '1', 'mutant must be refused'
report = json.loads((out / 'preflight.json').read_text())
assert report['native_mismatches'] == ['061_constEnumErrors'], report
assert not (out / 'results.json').exists(), 'refused binary was timed'
assert not list((out / 'raw').rglob('*-run-*.command.json')), 'timing began before refusal'
folder = out / 'raw/061_constEnumErrors'
node = (folder / 'node-preflight.stdout').read_bytes()
mutant = (folder / 'native-preflight.stdout').read_bytes()
assert mutant == node.replace(b'Enum declarations can only merge', b'MUTATED declarations can only merge', 1)
for suffix in ('stderr', 'exit'):
    assert (folder / ('node-preflight.' + suffix)).read_bytes() == (folder / ('native-preflight.' + suffix)).read_bytes()
print('PASS: one diagnostic changed; stderr and exit preserved; global preflight refused all timing')
PY
