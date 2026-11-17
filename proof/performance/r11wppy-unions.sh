# #r11wppy: every never[] stop left on the internal head (adaptations 95 and 96, this commit's parent), with its
# source line and the operators around its column, so the inferred-union sites can be classed by shape. Run after
# environment.sh, sealed-job.sh, sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; exit 3; }
"$BIN" "$A/src/compiler/sys.ts" > "$TMPDIR/scan.json" 2>/dev/null
echo "scan exit $?"
python3 - "$TMPDIR/scan.json" "$A" <<'PYTHON'
import json, re, sys
rows = json.load(open(sys.argv[1])).get('Stops') or []
root = sys.argv[2]
seen = set()
for s in rows:
    site = re.sub(r'^.*/adapted/', '', s['Site'])
    if 'never[]' not in s['Words'] or (site, s['Words']) in seen: continue
    seen.add((site, s['Words']))
    file, line, column = site.rsplit(':', 2)
    text = open(f"{root}/{file}", newline='').read().split('\n')
    n = int(line) - 1
    here = text[n].rstrip('\r')
    before, after = here[:int(column) - 1], here[int(column) - 1:]
    ops = [op for op in ('||', '??', '?', ':', 'return', '=', 'Object.entries', 'new Map', '.push(', 'concatenate', 'append(') if op in before[-60:] or op in after[:40]]
    print(f"STOP {site} | {s['Rule'].split('/')[-1]} | {s['Words'][:150]}")
    print(f"  {n}| {text[n - 1].rstrip(chr(13)).strip()[:180]}" if n > 0 else "")
    print(f"  {n + 1}> {here.strip()[:220]}")
    print(f"  ops near: {' '.join(ops)}")
print(f"TOTAL never[] distinct stops: {len(seen)}")
PYTHON
echo VERDICT DONE
