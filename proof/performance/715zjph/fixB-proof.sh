# #715zjph fix B: a checked access's view judgment kept per lowering. gofmt, vet, the count pin and its mutant,
# internal/lower, statecopy, stage3/fixtures and internal/oracle whole, and the casts-by-writes timing. Run after
# environment.sh, sandboxed-job.sh, sandboxed_npm stage3/api and sandboxed_products ./internal/lower ./internal/oracle
# ./stage3/fixtures. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/b-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/b-$name.log" | head -10 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/ stage3/census/latent/statecopy/)" ]'
step vet go vet ./internal/lower/
step pin go test -count=1 -v -run '^TestViewJudgmentKeptPerCheckedAccess$' ./internal/lower/
grep -E '_test.go:[0-9]+:' "$TMPDIR/b-pin.log" | head -2 | cut -c1-200
F=internal/lower/view_sites.go
cp "$F" "$TMPDIR/view_sites.go"
python3 - "$F" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p).read()
old = '\t\t\tif l.keepsViewJudgments() {\n\t\t\t\tl.keepViewSiteJudged(part, refused)\n\t\t\t}\n'
assert s.count(old) == 1
open(p, 'w').write(s.replace(old, ''))
PYTHON
go test -count=1 -v -run '^TestViewJudgmentKeptPerCheckedAccess$' ./internal/lower/ > "$TMPDIR/b-m.log" 2>&1
if grep -qE 'build failed|declared and not used' "$TMPDIR/b-m.log"; then echo "FAIL mutant doesn't compile"; bad=1
elif grep -q -- '--- FAIL: TestViewJudgmentKeptPerCheckedAccess' "$TMPDIR/b-m.log"; then echo "PASS mutant: the pin red"; grep -E '_test.go:[0-9]+:' "$TMPDIR/b-m.log" | head -1 | cut -c1-200
else echo "FAIL mutant: the pin stayed green"; bad=1; fi
cp "$TMPDIR/view_sites.go" "$F"
go test -count=1 -v -run '^TestProbeKindUnionProfile$' ./internal/lower/ 2>&1 | grep -E 'TIMING' | sed 's/^ *//' | cut -c1-160
step lower go test -count=1 -timeout 30m ./internal/lower/
step statecopy go test -count=1 -timeout 20m ./stage3/census/latent/...
step fixtures go test -count=1 -timeout 15m ./stage3/fixtures/...
start=$(date +%s); go test -count=1 -timeout 30m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1; echo "oracle exit $?, $(( $(date +%s) - start )) s"
other=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS oracle: only main's sealed set" || { echo "FAIL oracle: $other"; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
