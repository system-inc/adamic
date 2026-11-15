# #6xt6p1f's proof, sealed: bash proof/performance/6xt6p1f.sh focused|whole, after environment.sh and sealed-job.sh are
# sourced. Mutants restore by copy, since a sealed tree has no .git. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/ps-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/ps-$name.log" | head -8 | cut -c1-240; bad=1; }; }
pins='^(TestPlaceholderSlotsAreWalkedOncePerCheckerAndProgram|TestPlaceholderSlotsAreClonedInAndOut|TestPlaceholderSlotsUnderAMapperOrAKeptSlotLeaveTheCacheAlone|TestPlaceholderSlotsCachedAgreeWithAWalk|TestRefusalRulesReadTheirSlotThroughSlotType)$'
case "${1:-focused}" in
focused)
  step gofmt bash -c '[ -z "$(gofmt -l internal/lower/ stage3/census/latent/)" ]'
  step vet go vet ./internal/lower/ ./stage3/census/latent/...
  step pins go test -count=1 -v -run "$pins" ./internal/lower/
  grep -E '^(--- (PASS|FAIL|SKIP)|ok|FAIL)' "$TMPDIR/ps-pins.log" | head -8
  mutant() { local label=$1 old=$2 new=$3 test=$4 f=internal/lower/placeholder.go
    cp $f "$TMPDIR/m.orig"
    python3 -I - "$f" "$old" "$new" <<'PY' || { echo "FAIL mutant $label didn't apply"; bad=1; cp "$TMPDIR/m.orig" $f; return; }
import sys
path, old, new = sys.argv[1:]
text = open(path).read()
if text.count(old) != 1: sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
    go test -count=1 -v -run "^$test\$" ./internal/lower/ > "$TMPDIR/ps-m-$label.log" 2>&1
    cp "$TMPDIR/m.orig" $f
    if grep -q -E 'build failed|\[setup failed\]|declared and not used|imported and not used' "$TMPDIR/ps-m-$label.log"; then echo "FAIL mutant $label didn't compile"; grep -E '\.go:[0-9]+:' "$TMPDIR/ps-m-$label.log" | head -2; bad=1
    elif grep -q -E "^--- FAIL: $test " "$TMPDIR/ps-m-$label.log"; then echo "PASS mutant $label red: $test"; grep -m1 -E '_test.go:[0-9]+:|panic: placeholder' "$TMPDIR/ps-m-$label.log" | cut -c1-200
    elif grep -q -E "^--- SKIP: $test " "$TMPDIR/ps-m-$label.log"; then echo "FAIL mutant $label: $test skipped"; bad=1
    else echo "FAIL mutant $label survived: $test"; bad=1; fi; }
  mutant cache-dropped 'if l.placeholderSlots == nil || l.typeMapper != nil' 'if true || l.placeholderSlots == nil || l.typeMapper != nil' TestPlaceholderSlotsAreWalkedOncePerCheckerAndProgram
  mutant taken-by-reference 'l.result.PlaceholderSources, l.result.PlaceholderViews = maps.Clone(answer.sources), maps.Clone(answer.views)' 'l.result.PlaceholderSources, l.result.PlaceholderViews = answer.sources, maps.Clone(answer.views)' TestPlaceholderSlotsAreClonedInAndOut
  mutant guard-dropped 'if l.placeholderSlots == nil || l.typeMapper != nil || len(l.slots.explicit) != 0 {' 'if l.placeholderSlots == nil {' TestPlaceholderSlotsUnderAMapperOrAKeptSlotLeaveTheCacheAlone
  mutant stored-empty 'placeholderSlotAnswer{sources: maps.Clone(l.result.PlaceholderSources), views:' 'placeholderSlotAnswer{sources: ir.SortedMap[string, string]{}, views:' TestPlaceholderSlotsCachedAgreeWithAWalk
  step census go test -count=1 ./stage3/census/latent/...
  ;;
whole)
  step lower go test -count=1 -timeout 40m ./internal/lower/
  ;;
esac
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
