# #r11wppy item 2's proof, sealed: the parent of this commit is the head. Run after environment.sh, sealed-job.sh,
# sealed_git and sealed_products ./internal/lower ./stage3/fixtures/... Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/f-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/f-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
step vet go vet ./internal/lower/
step pins go test -count=1 -timeout 30m -v -run '^TestFreshAssignment' ./internal/lower/
grep -E '^ *--- (PASS|FAIL)' "$TMPDIR/f-pins.log" | sed 's/^ *//' | cut -c1-120 | head -45
go test -count=1 -timeout 10m -v -run '^TestZZChainJavaScriptProbe$' ./internal/lower/ 2>&1 | grep -E '^PROBE|^--- |^FAIL|_test.go' | cut -c1-400 | head -120
grep -E ' stops: ' "$TMPDIR/f-pins.log" | sed 's/^ *//' | cut -c1-300 | head -12
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
	go test -count=1 -timeout 30m -v -run '^TestFreshAssignment' ./internal/lower/ > "$TMPDIR/f-mutant.log" 2>&1
	if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/f-mutant.log"; then echo "FAIL mutant doesn't compile"; bad=1; else
		red=$(grep -cE '^ *--- FAIL: TestFreshAssignmentHandlesJudgeEveryHandle ' "$TMPDIR/f-mutant.log")
		others=$(grep -E '^ *--- FAIL: TestFreshAssignment' "$TMPDIR/f-mutant.log" | grep -v 'TestFreshAssignmentHandlesJudgeEveryHandle' | grep -vc 'TestFreshAssignmentValueLowers')
		echo "$([ "$red" = 1 ] && [ "$others" = 0 ] && echo PASS || echo FAIL) mutant: the direct pin $([ "$red" = 1 ] && echo red || echo green), $others others red"; { [ "$red" = 1 ] && [ "$others" = 0 ]; } || bad=1
		grep -E 'judge_every|JudgeEveryHandle.*_test.go|_test.go:[0-9]+: got map' "$TMPDIR/f-mutant.log" | head -2 | cut -c1-300
	fi
fi
cp "$TMPDIR/f-original" "$F"
# This round reads the chain probe and the mutant; internal/lower whole and stage3/fixtures run once the chain is understood.
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
