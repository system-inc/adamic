# #hjspb3b: collectViewedElements once per (checker, program) in the latent census. gofmt, vet, the cache's tests,
# the count mutant, internal/lower and the latent census's packages whole. Run after environment.sh, sandboxed-job.sh,
# sandboxed_typescript_source, sandboxed_npm stage3/api and sandboxed_products ./internal/lower. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/h-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/h-$name.log" | head -10 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/ stage3/census/latent/statecopy/)" ]'
step vet go vet ./internal/lower/
step pins go test -count=1 -v -run '^TestViewedElements' ./internal/lower/
grep -E '^--- (PASS|FAIL)' "$TMPDIR/h-pins.log" | cut -c1-120
F=internal/lower/viewed_elements_cache.go
cp "$F" "$TMPDIR/cache.go"
python3 - "$F" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p).read(); old = 'l.viewedElements.answers[key] = l.viewedElementsTaken(modules, stop)'
assert s.count(old) == 1
open(p, 'w').write(s.replace(old, '_ = key'))
PYTHON
go test -count=1 -v -run '^TestViewedElementsAreWalkedOncePerCheckerAndProgram$' ./internal/lower/ > "$TMPDIR/h-m.log" 2>&1
if grep -qE 'build failed|declared and not used' "$TMPDIR/h-m.log"; then echo "FAIL mutant doesn't compile"; grep -E '\.go:[0-9]+' "$TMPDIR/h-m.log" | head -3; bad=1
elif grep -q -- '--- FAIL: TestViewedElementsAreWalkedOncePerCheckerAndProgram' "$TMPDIR/h-m.log"; then echo "PASS mutant: the count pin red"; grep -E '_test.go:[0-9]+:' "$TMPDIR/h-m.log" | head -1 | cut -c1-200
else echo "FAIL mutant: the count pin stayed green"; bad=1; fi
cp "$TMPDIR/cache.go" "$F"
step lower go test -count=1 -timeout 30m ./internal/lower/
step census go test -count=1 -timeout 30m ./stage3/census/latent/...
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
