# #k8ed1f2's arrow-at-site rule, sealed: the parent of this commit is the head. gofmt, vet, its pins, six mutants
# each with the tests it must red, the this-parameter probe, internal/lower whole and stage3/fixtures. Run after
# environment.sh, sandboxed-job.sh, sandboxed_npm stage3/api and sandboxed_products ./internal/lower ./stage3/fixtures.
# Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/a-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/a-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
step vet go vet ./internal/lower/
pins='^(TestFreshArrow|TestFreshMap|TestFreshFunction|TestFreshValue)'
step pins go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/
grep -E '^--- (PASS|FAIL)' "$TMPDIR/a-pins.log" | cut -c1-120
go test -count=1 -v -run '^TestProbeLiteralThisParameter$' ./internal/lower/ 2>&1 | grep -E 'PROBE' | cut -c1-400
F=internal/lower/invariance.go
cp "$F" "$TMPDIR/a-original"
mutant() {
	local name=$1 old=$2 new=$3 want=$4
	cp "$TMPDIR/a-original" "$F"
	python3 - "$F" "$old" "$new" <<'PYTHON'
import sys
p, old, new = sys.argv[1:4]; s = open(p).read()
assert s.count(old) == 1, 'mutant anchor: ' + old
open(p, 'w').write(s.replace(old, new))
PYTHON
	if cmp -s "$F" "$TMPDIR/a-original"; then echo "FAIL mutant $name didn't apply"; bad=1; return; fi
	go test -count=1 -timeout 20m -v -run "$pins" ./internal/lower/ > "$TMPDIR/a-m-$name.log" 2>&1
	if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/a-m-$name.log"; then echo "FAIL mutant $name doesn't compile"; grep -E '\.go:[0-9]+' "$TMPDIR/a-m-$name.log" | head -3; bad=1; return; fi
	local reds=$(grep -oE '^--- FAIL: Test[A-Za-z]+' "$TMPDIR/a-m-$name.log" | cut -d' ' -f3 | sort -u | tr '\n' ' ')
	[ "$reds" = "$want " ] && echo "PASS mutant $name: red $reds" || { echo "FAIL mutant $name: red [$reds], want [$want]"; bad=1; }
}
mutant madeAtSite '		if !l.madeAtSite(returned) {' '		if false {' 'TestFreshMapNullishOfALiteralStaysRefused'
mutant lendsItself '		if lendsItself(returned) {' '		if false {' 'TestFreshArrowLiteralLendingItselfStaysRefused'
mutant literalJudged '		for _, member := range l.literalSlotMembers(value, target) {' '		for _, member := range l.literalSlotMembers(value, target)[:0] {' 'TestFreshArrowLiteralHoldingAHeldArrayStaysRefused TestFreshMapLiteralHoldingAHeldArrayStaysRefused'
mutant arrowArm '		if found, judged := l.freshFunctionWidened(node, own, contextual); judged {' '		if found, judged := l.freshFunctionWidened(node, own, contextual); judged && false {' 'TestFreshArrowAssignedHoldingAHeldValueLowers TestFreshArrowExpressionBodyLowers TestFreshArrowTryFinallyLowers'
mutant mapArm 'return l.freshContainerWidened(own, contextual, l.freshMapCallback(node), visited)' 'return l.freshContainerWidened(own, contextual, nil, visited)' 'TestFreshMapEmptyArrayElementsLower TestFreshMapLiteralElementsLower'
mutant namesItself '	if function.Kind == ast.KindFunctionExpression && namesItself(function) {' '	if function.Kind == ast.KindFunctionExpression && false && namesItself(function) {' 'TestFreshFunctionStoringItselfStaysRefused'
cp "$TMPDIR/a-original" "$F"
step lower go test -count=1 -timeout 30m ./internal/lower/
step fixtures go test -count=1 -timeout 15m ./stage3/fixtures/...
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
