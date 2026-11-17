# #96pc420's bound, sealed: the parent of this commit is the head. Its pins, two mutants, the packages the
# rule reaches, and the admission delta. Run after environment.sh, sealed-job.sh, sealed_git, sealed_npm
# stage3/api and sealed_products ./internal/lower ./internal/oracle ./stage3/fixtures. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/b-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/b-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
step vet go vet ./internal/lower/
pins='^TestPromiseIntoVoid'
step pins go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/
grep -E '^ *--- (PASS|FAIL)' "$TMPDIR/b-pins.log" | sed 's/^ *//' | cut -c1-120
F=internal/lower/promise_into_void.go
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
	local reds=$(grep -oE '^--- FAIL: TestPromiseIntoVoid[A-Za-z]+' "$TMPDIR/b-m-$name.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
	[ "$reds" = "$want " ] && echo "PASS mutant $name: red $reds" || { echo "FAIL mutant $name: red [$reds], want [$want]"; bad=1; }
}
mutant bound64 'const promiseIntoVoidDepth = 256' 'const promiseIntoVoidDepth = 64' TestPromiseIntoVoidPassesPastSixtyFour
mutant noseen '	if seen[pair] {
		return nil
	}
	seen[pair] = true
	if depth >= promiseIntoVoidDepth' '	seen[pair] = true
	if depth >= promiseIntoVoidDepth' TestPromiseIntoVoidEndsACycleWithoutTheBound
cp "$TMPDIR/b-original" "$F"
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 20m ./stage3/fixtures/...
start=$(date +%s)
go test -count=1 -timeout 40m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1
echo "oracle exit $?, $(( $(date +%s) - start )) s"
fails=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
other=$(echo "$fails" | tr ' ' '\n' | grep -v '^$' | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS oracle: only main's sealed set" || { echo "FAIL oracle beyond main's sealed set: $other"; bad=1; }
start=$(date +%s)
python3 cloud/admission-corpus/manifest.py --sha HEAD~1 > "$TMPDIR/manifest.json" 2> "$TMPDIR/manifest.err" || { echo "FAIL manifest"; tail -5 "$TMPDIR/manifest.err"; bad=1; }
go run ./cmd/adamic-admission-delta --base HEAD~2 --head HEAD~1 --manifest "$TMPDIR/manifest.json" --manifest-generator cloud/admission-corpus/manifest.py --json > "$TMPDIR/delta.json" 2> "$TMPDIR/delta.err"
code=$?
echo "$([ $code = 0 ] && echo PASS || echo FAIL) admission delta exit $code, $(( $(date +%s) - start )) s"
tail -4 "$TMPDIR/delta.err" | cut -c1-300
python3 - "$TMPDIR/delta.json" <<'PYTHON'
import json, sys
try:
    d = json.load(open(sys.argv[1]))
except Exception as e:
    print(f"delta json unreadable: {e}"); sys.exit(0)
for key in ('summary', 'counts', 'newly_admitted', 'newly_refused', 'disagreements', 'void'):
    if key in d:
        v = d[key]
        print(f"  {key}: {json.dumps(v)[:400] if not isinstance(v, list) else str(len(v)) + ' ' + json.dumps(v[:5])[:400]}")
PYTHON
[ $code = 0 ] || bad=1
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
