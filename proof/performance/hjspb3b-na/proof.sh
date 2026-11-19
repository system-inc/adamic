# #hjspb3b follow-ups: noteInheritance and noteAccessorNames once per run in the latent census. gofmt, vet, both
# caches' tests, both count mutants, internal/lower whole, the latent packages, and the census overlay's in-process
# tests. Run after environment.sh, sandboxed-job.sh, sandboxed_typescript_source, sandboxed_npm stage3/api and
# sandboxed_products ./internal/lower. Never merged.
set -u
W=$(pwd); O=$TMPDIR/na; mkdir -p "$O"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$O/$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*\.go:[0-9]+:)' "$O/$name.log" | head -10 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/ stage3/census/latent/statecopy/)" ]'
step vet go vet ./internal/lower/
step pins go test -count=1 -v -run '^Test(Inheritance|AccessorNames)' ./internal/lower/
grep -E '^--- (PASS|FAIL)' "$O/pins.log" | cut -c1-120
# mutant <name> <file> <old> <test regexp>
mutant() {
  local name=$1 file=$2 old=$3 tests=$4
  cp "$file" "$O/$name.orig"
  python3 - "$file" "$old" <<'PYTHON'
import sys
p, old = sys.argv[1], sys.argv[2]; s = open(p).read()
assert s.count(old) == 1, old
open(p, 'w').write(s.replace(old, '_ = key'))
PYTHON
  go test -count=1 -v -run "$tests" ./internal/lower/ > "$O/m-$name.log" 2>&1
  if grep -qE 'build failed|declared and not used|imported and not used' "$O/m-$name.log"; then echo "FAIL mutant $name doesn't compile"; grep -E '\.go:[0-9]+' "$O/m-$name.log" | head -3; bad=1
  elif grep -q -- '--- FAIL: ' "$O/m-$name.log"; then echo "PASS mutant $name: red"; grep -E '^--- FAIL|_test.go:[0-9]+:' "$O/m-$name.log" | head -3 | cut -c1-200
  else echo "FAIL mutant $name: the pins stayed green"; bad=1; fi
  cp "$O/$name.orig" "$file"
}
mutant inheritance internal/lower/inheritance_cache.go 'l.inheritance.answers[key] = inheritanceAnswer{modules: slices.Clone(modules), derivedAncestors: maps.Clone(l.derivedAncestors)}' '^TestInheritance'
mutant accessors internal/lower/accessor_names_cache.go 'l.accessorNamesHeld.answers[key] = accessorNamesAnswer{modules: slices.Clone(modules), accessorNames: maps.Clone(l.accessorNames)}' '^TestAccessorNames'
git diff --quiet HEAD -- internal || { echo "FAIL a mutant wasn't restored"; bad=1; }
step lower go test -count=1 -timeout 30m ./internal/lower/
step latent go test -count=1 -timeout 30m ./stage3/census/latent/...
step overlay python3 stage3/census/latent/make_overlay.py "$W" "$O/latent"
gofmt -w "$O/latent/"*.go
step tool-build go build -buildvcs=false -overlay="$O/latent/overlay.json" -o "$O/census" ./stage3/census/latent/tool
step tool go test -count=1 -timeout 30m -overlay="$O/latent/overlay.json" ./stage3/census/latent/tool
go test -count=1 -timeout 30m -tags=census_overlay -overlay="$O/latent/overlay.json" ./stage3/census/latent > "$O/census-tag.log" 2>&1
other=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+' "$O/census-tag.log" | cut -d' ' -f3 | sort -u | grep -vE '^(TestLatentWrittenAnyDeferredCensusStopsZb0q0zd|TestLatentScopedTailEntryFindingLoom03dg)$' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS census-tag: only main's two known reds ($(grep -c '^--- FAIL' "$O/census-tag.log") failing)" || { echo "FAIL census-tag: $other"; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
