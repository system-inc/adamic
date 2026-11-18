# #k8ed1f2 changes A and B, sealed: the parent of this commit is the head. Its pins, two mutants, the packages
# they reach, and the census of sys.ts's closure before and after. Run after environment.sh, sealed-job.sh, sealed_git, sealed_npm
# stage3/api and sealed_products ./internal/lower ./internal/oracle ./stage3/fixtures. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/b-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/b-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
step vet go vet ./internal/lower/
pins='^TestFreshValue'
step pins go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/
grep -E '^ *--- (PASS|FAIL)' "$TMPDIR/b-pins.log" | sed 's/^ *//' | cut -c1-120
F=internal/lower/invariance.go
cp "$F" "$TMPDIR/b-original"
mutant() {
	local name=$1 old=$2 new=$3 want=$4
	cp "$TMPDIR/b-original" "$F"
	python3 - "$F" "$old" "$new" <<'PYTHON'
import sys
p, old, new = sys.argv[1:4]; s = open(p).read()
assert s.count(old) == 1, 'mutant anchor: ' + old
open(p, 'w').write(s.replace(old, new))
PYTHON
	if cmp -s "$F" "$TMPDIR/b-original"; then echo "FAIL mutant $name didn't apply"; bad=1; return; fi
	go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/ > "$TMPDIR/b-m-$name.log" 2>&1
	if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/b-m-$name.log"; then echo "FAIL mutant $name doesn't compile"; bad=1; return; fi
	local reds=$(grep -oE '^--- FAIL: TestFreshValue[A-Za-z]+' "$TMPDIR/b-m-$name.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
	[ "$reds" = "$want " ] && echo "PASS mutant $name: red $reds" || { echo "FAIL mutant $name: red [$reds], want [$want]"; bad=1; }
}
mutant optional '			if access.QuestionDotToken != nil {' '			if false {' TestFreshValueOptionalCopyLowers
mutant andarm '		case ast.KindAmpersandAmpersandToken:
			// Only the right is the value' '		case ast.KindCommaToken:
			// Only the right is the value' TestFreshValueAndRightLiteralLowers
cp "$TMPDIR/b-original" "$F"
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 20m ./stage3/fixtures/...
start=$(date +%s)
go test -count=1 -timeout 40m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1
echo "oracle exit $?, $(( $(date +%s) - start )) s"
fails=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
other=$(echo "$fails" | tr ' ' '\n' | grep -v '^$' | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS oracle: only main's sealed set" || { echo "FAIL oracle beyond main's sealed set: $other"; bad=1; }
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "FAIL apply"; bad=1; }
for side in after before; do
	if [ $side = before ]; then git show HEAD~1:internal/lower/invariance.go > internal/lower/invariance.go; fi
	go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product-$side.log" 2>&1 || { echo "FAIL product $side"; bad=1; continue; }
	BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product-$side.log" | tail -1 | cut -d' ' -f2)
	cp "$BIN" "$TMPDIR/probe-$side"
	"$TMPDIR/probe-$side" "$A/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2>/dev/null
	echo "SCAN $side exit $?"
done
cp "$TMPDIR/b-original" "$F"
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" <<'PYTHON'
import json, re, sys
def load(p):
    rows = json.load(open(p)).get('Stops') or []
    return set((s['Rule'], re.sub(r'^.*/adapted/', '', s['Site']), s['Words']) for s in rows)
b, a = load(sys.argv[1]), load(sys.argv[2])
gone, new = sorted(b - a), sorted(a - b)
print(f"before: distinct {len(b)}; after: distinct {len(a)}; before to after: {len(gone)} gone, {len(new)} new")
for x in gone: print(f"  - {x[1]}: {x[2][:150]}")
for x in new: print(f"  + {x[1]}: {x[0]} {x[2][:160]}")
PYTHON
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
