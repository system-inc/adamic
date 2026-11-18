# #k8ed1f2: the scan-all census for adaptation 99-readonly-family3-views, one side (with or without it) over the
# closures of the files named after the side. Prints each file's boundary and every stop as "STOP <site>: <rule>
# <words>". Run after environment.sh, sandboxed-job.sh, sandboxed_typescript_source and sandboxed_npm stage3/api.
# Never merged.
set -u
side=$1; shift
D=stage3/adapt/99-readonly-family3-views
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg), side $side, files $*"
[ "$side" = without ] && mv "$D" "$TMPDIR/aside"
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "VERDICT FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }
[ "$side" = without ] && mv "$TMPDIR/aside" "$D"
go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product.log" 2>&1 || { echo "VERDICT FAIL product"; grep -E 'FAIL|\.go:[0-9]+' "$TMPDIR/product.log" | head -5; exit 1; }
BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product.log" | tail -1 | cut -d' ' -f2)
start=$(date +%s)
bad=0
for file in "$@"; do
	"$BIN" "$A/src/compiler/$file" > "$TMPDIR/$(echo $file | tr / _).json" 2>/dev/null
	code=$?
	echo "SCAN $side $file exit $code, $(( $(date +%s) - start )) s"
	[ $code = 0 ] || bad=1
done
python3 - "$TMPDIR" <<'PYTHON'
import json, re, sys, glob
stops = set()
for p in glob.glob(sys.argv[1] + '/*.ts.json'):
    report = json.load(open(p))
    print(f"BOUNDARY {p.rsplit('/', 1)[-1]}: {str(report.get('Boundary') or '')[:300]}")
    for s in report.get('Stops') or []:
        stops.add((re.sub(r'^.*/adapted/', '', s['Site']), s['Rule'], s['Words']))
print(f"stops: {len(stops)}")
for site, rule, words in sorted(stops): print(f"STOP {site}: {rule} {words[:160]}")
PYTHON
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
