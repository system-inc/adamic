# #r11wppy class 1's operand-landing rule, sealed: the parent of this commit is the head. Its pins, two mutants,
# the packages it reaches, and the census of sys.ts's closure with the rule and without it. Run after
# environment.sh, sealed-job.sh, sealed_git, sealed_typescript_source, sealed_npm stage3/api and sealed_products
# ./internal/lower ./internal/oracle ./stage3/fixtures. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/l-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/l-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
step vet go vet ./internal/lower/
pins='^TestOperandLanding'
step pins go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/
grep -E '^ *--- (PASS|FAIL)' "$TMPDIR/l-pins.log" | sed 's/^ *//' | cut -c1-120
F=internal/lower/invariance.go
cp "$F" "$TMPDIR/l-original"
mutant() {
	local name=$1 old=$2 new=$3 want=$4
	cp "$TMPDIR/l-original" "$F"
	python3 - "$F" "$old" "$new" <<'PYTHON'
import sys
p, old, new = sys.argv[1:4]; s = open(p).read()
assert s.count(old) == 1, 'mutant anchor: ' + old
open(p, 'w').write(s.replace(old, new))
PYTHON
	if cmp -s "$F" "$TMPDIR/l-original"; then echo "FAIL mutant $name didn't apply"; bad=1; return; fi
	go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/ > "$TMPDIR/l-m-$name.log" 2>&1
	if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/l-m-$name.log"; then echo "FAIL mutant $name doesn't compile"; bad=1; return; fi
	local reds=$(grep -oE '^ *--- FAIL: TestOperandLanding[A-Za-z]+/[^ ]+' "$TMPDIR/l-m-$name.log" | sed 's/^ *--- FAIL: //' | sort -u | tr '\n' ' ')
	echo "mutant $name: red [$reds], want rows matching $want"
	echo "$reds" | grep -qE "$want" || { echo "FAIL mutant $name: no red row matches $want"; bad=1; }
}
mutant unannotated-read '		declared = parent.Initializer() == value && parent.Type() != nil' '		if parent.Initializer() == value && parent.Type() == nil {
			return nil, true
		}
		declared = parent.Initializer() == value && parent.Type() != nil' 'TestOperandLandingRefused/'
mutant read-walks-nothing '	source, target := l.withoutUndefined(own), l.withoutUndefined(contextual)
	for _, viewed := range l.containers(target) {' '	source, target := l.withoutUndefined(own), l.withoutUndefined(contextual)
	if source != nil {
		return nil
	}
	for _, viewed := range l.containers(target) {' 'TestOperandLandingRefused/'
cp "$TMPDIR/l-original" "$F"
grep -E 'FAIL: TestOperandLanding' "$TMPDIR/l-m-unannotated-read.log" "$TMPDIR/l-m-read-walks-nothing.log" | sed 's#^.*/l-m-##' | cut -c1-160
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 20m ./stage3/fixtures/...
start=$(date +%s)
go test -count=1 -timeout 40m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1
echo "oracle exit $?, $(( $(date +%s) - start )) s"
fails=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
other=$(echo "$fails" | tr ' ' '\n' | grep -v '^$' | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS oracle: only main's sealed set" || { echo "FAIL oracle beyond main's sealed set: $other"; bad=1; }
# The census: sys.ts's closure with the rule (after) and with invariance.go as on main (before).
export SCAN_ALL_SCRATCH="$TMPDIR/scan"; mkdir -p "$SCAN_ALL_SCRATCH"
A="$SCAN_ALL_SCRATCH/adapted"
bash stage3/apply.sh --build-product "$A" > "$TMPDIR/apply.log" 2>&1 || { echo "FAIL apply"; bad=1; }
for side in after before; do
	if [ $side = before ]; then git show HEAD~2:internal/lower/invariance.go > internal/lower/invariance.go; fi
	go test -v -count=1 -timeout=20m -run '^TestProduct_ScanAllInstrumentedProbe$' ./stage3/census/scan-all > "$TMPDIR/product-$side.log" 2>&1 || { echo "FAIL product $side"; bad=1; continue; }
	BIN=$(grep -oE 'probe: /[^ ]+' "$TMPDIR/product-$side.log" | tail -1 | cut -d' ' -f2)
	cp "$BIN" "$TMPDIR/probe-$side"
	"$TMPDIR/probe-$side" "$A/src/compiler/sys.ts" > "$TMPDIR/$side.json" 2>/dev/null
	echo "SCAN $side exit $?"
done
cp "$TMPDIR/l-original" "$F"
cmp -s "$TMPDIR/probe-after" "$TMPDIR/probe-before" && echo "WARN the two probes are the same binary"
python3 - "$TMPDIR/before.json" "$TMPDIR/after.json" <<'PYTHON'
import json, re, sys
def load(p):
    rows = json.load(open(p)).get('Stops') or []
    return set((s['Rule'], re.sub(r'^.*/adapted/', '', s['Site']), s['Words']) for s in rows)
b, a = load(sys.argv[1]), load(sys.argv[2])
gone, new = sorted(b - a), sorted(a - b)
print(f"before: distinct {len(b)}, never[] {sum(1 for x in b if 'never[]' in x[2])}; after: distinct {len(a)}, never[] {sum(1 for x in a if 'never[]' in x[2])}")
print(f"before to after: {len(gone)} gone, {len(new)} new")
for x in gone: print(f"  - {x[1]}: {x[2][:140]}")
for x in new: print(f"  + {x[1]}: {x[0]} {x[2][:160]}")
PYTHON
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
