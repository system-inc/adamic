# #k8ed1f2: objects/native.json regenerated and compared, its mutant, then internal/oracle whole. Run after
# environment.sh, sandboxed-job.sh, sandboxed_npm stage3/api and sandboxed_products ./internal/oracle. Never merged.
set -u
bad=0
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
N=stage3/fixtures/objects/native.json
cp "$N" "$TMPDIR/native.json"
go test -count=1 -timeout 20m -run '^TestObjectFixtureNativeSummaryMatchesCompiledCases$' ./internal/oracle/ -args -update-graph-object-native > "$TMPDIR/n.log" 2>&1
code=$?
if [ $code = 0 ] && cmp -s "$N" "$TMPDIR/native.json"; then echo "PASS native.json regenerates byte-identical"; else echo "FAIL native.json exit $code"; diff "$TMPDIR/native.json" "$N" | head -20; grep -E 'FAIL|_test.go:' "$TMPDIR/n.log" | head -5 | cut -c1-300; bad=1; fi
cp "$TMPDIR/native.json" "$N"
F=internal/lower/invariance.go
cp "$F" "$TMPDIR/inv"
python3 - "$F" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p).read(); old = '			if access.QuestionDotToken != nil {'
assert s.count(old) == 1
open(p, 'w').write(s.replace(old, '			if false {'))
PYTHON
go test -count=1 -timeout 20m -run '^TestObjectFixtureNativeSummaryMatchesCompiledCases$' ./internal/oracle/ -args -update-graph-object-native > "$TMPDIR/nm.log" 2>&1
code=$?
cp "$TMPDIR/inv" "$F"; cp "$TMPDIR/native.json" "$N"
if [ $code != 0 ] && grep -q '20_queue_optional_call' "$TMPDIR/nm.log"; then echo "PASS mutant native20: red naming 20_queue_optional_call"; else echo "FAIL mutant native20: exit $code"; grep -E 'FAIL|_test.go:' "$TMPDIR/nm.log" | head -5 | cut -c1-300; bad=1; fi
grep -E '_test.go:' "$TMPDIR/nm.log" | head -2 | cut -c1-300
bash proof/performance/k8ed1f2/arrow-oracle.sh || bad=1
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
