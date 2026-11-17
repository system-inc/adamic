# #96pc420: which stops take the place of the 24 bound stops when the bound is 256. Scans sys.ts's closure at
# bound 64 and at 256 with the env-gated probe in this commit's parent, and lists every distinct stop that
# differs. Run after environment.sh, sealed-job.sh, sealed_typescript_source and sealed_npm stage3/api. Never merged.
set -u
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; exit 3; }
for bound in 64 256; do
	ADAMIC_PROBE_PROMISE_BOUND=$bound ADAMIC_PROBE_PROMISE_TIME="$TMPDIR/t$bound" "$BIN" "$A/src/compiler/sys.ts" > "$TMPDIR/b$bound.json" 2>/dev/null
	echo "bound $bound exit $?"
done
python3 - "$TMPDIR/b64.json" "$TMPDIR/b256.json" <<'PYTHON'
import json, re, sys
def load(p):
    rows = json.load(open(p)).get('Stops') or []
    return set((s['Rule'], re.sub(r'^.*/adapted/', '', s['Site']), s['Words']) for s in rows)
a, b = load(sys.argv[1]), load(sys.argv[2])
gone, new = sorted(a - b), sorted(b - a)
print(f"64 to 256: {len(gone)} gone, {len(new)} new")
for x in gone: print(f"  - {x[1]}: {x[0]}: {x[2][:120]}")
for x in new: print(f"  + {x[1]}: {x[0]}: {x[2][:200]}")
PYTHON
echo VERDICT DONE
