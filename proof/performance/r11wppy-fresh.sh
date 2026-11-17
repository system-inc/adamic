# #r11wppy item 2's proof, sealed: the parent of this commit is the head. Run after environment.sh, sealed-job.sh,
# sealed_git and sealed_products ./internal/lower ./stage3/fixtures/... Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/f-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/f-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
step vet go vet ./internal/lower/
step pins go test -count=1 -timeout 30m -v -run '^TestFreshAssignmentValue' ./internal/lower/
grep -E '^ *--- (PASS|FAIL)' "$TMPDIR/f-pins.log" | sed 's/^ *//' | head -30
# Mutant: only the outermost handle is judged.
F=internal/lower/assignment_values.go
cp "$F" "$TMPDIR/f-original"
python3 - "$F" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p).read()
old = "\t\t\treturn false\n\t\t}\n\t}\n\treturn true\n}\n\n// assignmentPrimitiveMembers"
new = "\t\t\treturn false\n\t\t}\n\t\tbreak\n\t}\n\treturn true\n}\n\n// assignmentPrimitiveMembers"
assert s.count(old) == 1, 'mutant anchor'
open(p, 'w').write(s.replace(old, new))
PYTHON
if cmp -s "$F" "$TMPDIR/f-original"; then echo "FAIL mutant didn't apply"; bad=1; else
	go test -count=1 -timeout 30m -v -run '^TestFreshAssignmentValue' ./internal/lower/ > "$TMPDIR/f-mutant.log" 2>&1
	if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/f-mutant.log"; then echo "FAIL mutant doesn't compile"; bad=1; else
		reds=$(grep -cE '^ *--- FAIL: TestFreshAssignmentValueRefused/disagreeing_handles\.(a|ts) ' "$TMPDIR/f-mutant.log")
		others=$(grep -E '^ *--- FAIL: TestFreshAssignmentValue' "$TMPDIR/f-mutant.log" | grep -vcE 'disagreeing_handles|^ *--- FAIL: TestFreshAssignmentValueRefused \(')
		echo "$([ "$reds" = 2 ] && [ "$others" = 0 ] && echo PASS || echo FAIL) mutant: $reds of 2 disagreeing pins red, $others others"; { [ "$reds" = 2 ] && [ "$others" = 0 ]; } || bad=1
		grep -E 'disagreeing.*_test.go:[0-9]+:|_test.go:[0-9]+: got' "$TMPDIR/f-mutant.log" | head -2 | cut -c1-300
	fi
fi
cp "$TMPDIR/f-original" "$F"
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 30m ./stage3/fixtures/...
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
