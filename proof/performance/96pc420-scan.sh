# #96pc420's probe, sealed: the parent of this commit carries promise_probe.go. Builds scan-all's instrumented
# probe and tsc's adapted tree, then scans sys.ts's closure once per variant named in $@ (base, chain, b256,
# b4096, reach), one at a time on one box. Run after environment.sh, sealed-job.sh, sealed_typescript_source
# and sealed_npm stage3/api. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
start=$(date +%s)
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; tail -20 "$TMPDIR/product.log"; exit 3; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
[ -x "$BIN" ] || { echo "VERDICT FAIL no probe binary"; grep -n probe "$TMPDIR/product.log" | tail -5; exit 3; }
echo "product $BIN, $(( $(date +%s) - start )) s"
A="$SCAN_ALL_SCRATCH/adapted"
start=$(date +%s)
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -8 "$TMPDIR/apply.log"; exit 3; }
echo "apply $(( $(date +%s) - start )) s"
ROOT="$A/src/compiler/sys.ts"
for variant in "$@"; do
	unset ADAMIC_PROBE_PROMISE_BOUND ADAMIC_PROBE_PROMISE_REACH ADAMIC_PROBE_PROMISE_CHAIN
	export ADAMIC_PROBE_PROMISE_TIME="$TMPDIR/$variant.time"
	case $variant in
		base) ;;
		chain) export ADAMIC_PROBE_PROMISE_CHAIN="$TMPDIR/chain.txt" ;;
		b256) export ADAMIC_PROBE_PROMISE_BOUND=256 ;;
		b4096) export ADAMIC_PROBE_PROMISE_BOUND=4096 ;;
		reach) export ADAMIC_PROBE_PROMISE_REACH=1 ;;
	esac
	start=$(date +%s)
	"$BIN" "$ROOT" > "$TMPDIR/$variant.json" 2> "$TMPDIR/$variant.stderr"
	code=$?
	echo "VARIANT $variant exit $code, $(( $(date +%s) - start )) s wall"
	echo "  rule totals: $(cat "$TMPDIR/$variant.time" 2>/dev/null)"
	python3 - "$TMPDIR/$variant.json" <<'PYTHON'
import collections, json, sys
r = json.load(open(sys.argv[1]))
rows = r.get('Stops') or []
triples = set((s['Rule'], s['Site'], s['Words']) for s in rows)
mine = [s for s in rows if 'PromiseIntoVoid' in s['Rule']]
print(f"  events {len(rows)}, distinct {len(triples)}, completed {r.get('Completed')}, boundary {r.get('Boundary')}")
print(f"  refusePromiseIntoVoid: {len(mine)} events, {len(set((s['Site'], s['Words']) for s in mine))} distinct")
for words, count in collections.Counter(s['Words'] for s in mine).most_common(6):
    print(f"    {count}: {words[:220]}")
for s in mine:
    if 'deeper than' not in s['Words']:
        print(f"    discard at {s['Site']}: {s['Words'][:200]}")
PYTHON
done
if [ -s "$TMPDIR/chain.txt" ]; then
	echo "CHAINS: $(grep -c '^====' "$TMPDIR/chain.txt") stops"
	# Each stop's site and the last 24 pairs of its chain, so the tail shows what keeps it growing.
	awk '/^====/ { if (n) for (i = (n > 24 ? n - 24 : 0); i < n; i++) print lines[i]; print; n = 0; next } { lines[n++] = $0 } END { for (i = (n > 24 ? n - 24 : 0); i < n; i++) print lines[i] }' "$TMPDIR/chain.txt" | cut -c1-240 | head -400
fi
echo "VERDICT OK"
