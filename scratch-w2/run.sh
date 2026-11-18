#!/usr/bin/env bash
# Scratch for #38kdd4d's W2 cause, never lands: the per-field table of writes through Mutable views, on apply's tree.
set -o pipefail
root=$(pwd)
STAGE3_CACHE="$TMPDIR/c" bash stage3/apply.sh "$TMPDIR/t" > "$TMPDIR/apply.log" 2>&1 || { echo "FAIL apply"; tail -30 "$TMPDIR/apply.log"; exit 1; }
NODE_PATH="$root/stage3/api/node_modules" node scratch-w2/w2scout.cjs "$TMPDIR/t" scratch-w2/w2-rows-85.tsv "$TMPDIR/w2.json" || { echo "FAIL scout"; exit 1; }
echo "== rows"
python3 -I - "$TMPDIR/w2.json" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
rows = d['entries'] if isinstance(d.get('entries'), list) else d['rows']
for r in rows:
    fields = sorted({w['field'] for w in r.get('writes') or []})
    making = sorted({w['making'] for w in r.get('writes') or []})
    print('ROW', r.get('where'), '|', r.get('road'), '|', r.get('callee') or '-', '|', r.get('function') or '-', '|', ','.join(fields) or 'NONE', '|', ','.join(making) or '-', '|', 'fresh' if r.get('freshArgument') else '', '|', (r.get('expression') or '')[:90])
PY
echo PASS
