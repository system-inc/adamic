# #e1wpvc7: buildcache classes a read by the deepest of home and the temporary places. vet, the new test, the two tests
# that were red sealed, both mutants, buildcache whole and compilercensus. Prints HOME and TMPDIR so the box's nesting is
# on the record. Never merged.
set -u
O=${TMPDIR:-/tmp}/e1w; mkdir -p "$O"; bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$O/$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*\.go:[0-9]+:)' "$O/$name.log" | head -12 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg), HOME=$HOME TMPDIR=${TMPDIR:-unset}"
step vet go vet ./internal/buildcache/
step deepest go test -count=1 -v -run '^TestAPathIsClassedByTheDeepestOfHomeAndTheTemporaryPlaces$' ./internal/buildcache/ && ok=1 || ok=0
grep -E '^ +--- (PASS|FAIL)' "$O/deepest.log" | sed 's/^ *//' | cut -c1-140
step were-red go test -count=1 -v -run '^(TestAReadNoKeyCanNameIsRefused|TestAHandedBuildsAbsoluteReadUnderItsPathsRefusesIt)$' ./internal/buildcache/
grep -E '^--- (PASS|FAIL)' "$O/were-red.log" | cut -c1-140
F=internal/buildcache/settle.go
OLD='	if deepest >= 0 && !(under(name, facts.home) && len(strings.TrimSuffix(facts.home, "/")) >= deepest) {'
for m in first:'	if deepest >= 0 {' home:'	if deepest >= 0 && !under(name, facts.home) {'; do
  name=${m%%:*}; new=${m#*:}
  cp "$F" "$O/settle.orig"
  python3 - "$F" "$OLD" "$new" <<'PYTHON'
import sys
p, old, new = sys.argv[1:4]; s = open(p).read()
assert s.count(old) == 1
open(p, 'w').write(s.replace(old, new))
PYTHON
  go test -count=1 -v -run '^TestAPathIsClassedByTheDeepestOfHomeAndTheTemporaryPlaces$' ./internal/buildcache/ > "$O/m-$name.log" 2>&1
  if [ "$ok" != 1 ]; then echo "FAIL mutant $name: the test wasn't green first"; bad=1
  elif grep -qE 'build failed|declared and not used' "$O/m-$name.log"; then echo "FAIL mutant $name doesn't compile"; bad=1
  elif grep -q -- '--- FAIL: ' "$O/m-$name.log"; then echo "PASS mutant $name: red"; grep -E '^ +--- FAIL' "$O/m-$name.log" | sed 's/^ *//' | cut -c1-140
  else echo "FAIL mutant $name: stayed green"; bad=1; fi
  cp "$O/settle.orig" "$F"
done
step buildcache go test -count=1 -timeout 30m ./internal/buildcache/...
grep -E '^(ok|FAIL|---)' "$O/buildcache.log" | grep -v PASS | head -12 | cut -c1-160
step census go test -count=1 -timeout 20m ./internal/compilercensus/...
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
