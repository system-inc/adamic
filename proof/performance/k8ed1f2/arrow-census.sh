# #k8ed1f2: the scan-all census of the family 3 sites' files on three sides, base fbe5b72b (before A and B), A and B
# 47d03322, and the arrow head (this commit's parent), every stop that leaves or arrives. Run after environment.sh,
# sandboxed-job.sh, sandboxed_typescript_source and sandboxed_npm stage3/api. Never merged.
set -u
bad=0
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "FAIL apply"; tail -5 "$TMPDIR/apply.log"; exit 3; }
files="symbolWalker.ts transformers/declarations.ts moduleSpecifiers.ts utilities.ts moduleNameResolver.ts"
F=internal/lower/invariance.go
cp "$F" "$TMPDIR/head-invariance.go"
for side in base ab head; do
	case $side in
	base) git show fbe5b72ba115a8f3060d4b5229fd578574f4505f:$F > $F ;;
	ab) git show 47d03322a9163eda5c6428c1e01dad96e6ac2501:$F > $F ;;
	head) cp "$TMPDIR/head-invariance.go" $F ;;
	esac
	start=$(date +%s)
	go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product-$side.log" 2>&1 || { echo "FAIL product $side"; grep -E 'FAIL|\.go:[0-9]+' "$TMPDIR/product-$side.log" | head -5; bad=1; continue; }
	BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product-$side.log" | tail -1 | cut -d' ' -f2)
	cp "$BIN" "$TMPDIR/probe-$side"
	for file in $files; do
		"$TMPDIR/probe-$side" "$A/src/compiler/$file" > "$TMPDIR/$side-$(echo $file | tr / _).json" 2>/dev/null
		echo "SCAN $side $file exit $?, $(( $(date +%s) - start )) s"
	done
done
cp "$TMPDIR/head-invariance.go" $F
python3 - "$TMPDIR" $files <<'PYTHON'
import json, re, sys, os
d, files = sys.argv[1], sys.argv[2:]
def load(side, f):
    p = os.path.join(d, f"{side}-{f.replace('/', '_')}.json")
    try: rows = json.load(open(p)).get('Stops') or []
    except Exception as e: print(f"  unreadable {p}: {e}"); return set()
    return set((s['Rule'], re.sub(r'^.*/adapted/', '', s['Site']), s['Words']) for s in rows)
for a, b in (('base', 'ab'), ('ab', 'head')):
    gone, new = set(), set()
    for f in files:
        x, y = load(a, f), load(b, f)
        gone |= x - y; new |= y - x
    print(f"{a} to {b}: {len(gone)} gone, {len(new)} new")
    for s in sorted(gone): print(f"  - {s[1]}: {s[2][:150]}")
    for s in sorted(new): print(f"  + {s[1]}: {s[0]} {s[2][:160]}")
PYTHON
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
