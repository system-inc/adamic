# #58r9613's proof: bash proof/performance/58r9613.sh focused|whole. The parent of this commit is the head.
set -u
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/p58-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; tail -3 /tmp/p58-sub.log; exit 3; }
echo "head $(git rev-parse HEAD^) on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > /tmp/p58-$name.log 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' /tmp/p58-$name.log | head -12; bad=1; }; return $code; }
# mutant <name> <file> <expected reds> <run pattern> <red pattern> <sed expression>
mutant() {
  local name=$1 file=$2 want=$3 run=$4 red=$5 expression=$6
  cp "$file" /tmp/p58-original
  sed -i "$expression" "$file"
  if cmp -s "$file" /tmp/p58-original; then echo "FAIL mutant $name didn't apply"; bad=1; else
    go test -count=1 -v -run "$run" ./internal/lower/ > /tmp/p58-mutant-$name.log 2>&1
    if grep -qE 'build failed|declared and not used|\[setup failed\]' /tmp/p58-mutant-$name.log; then echo "FAIL mutant $name doesn't compile"; grep -E '\.go:[0-9]+:[0-9]+:' /tmp/p58-mutant-$name.log | head -3; bad=1; else
      local reds; reds=$(grep -cE "^--- FAIL: ($red) " /tmp/p58-mutant-$name.log)
      echo "$([ "$reds" = "$want" ] && echo PASS || echo FAIL) mutant $name: $reds of $want red"; [ "$reds" = "$want" ] || bad=1
      grep -E "^--- FAIL: " /tmp/p58-mutant-$name.log | head -6
      grep -E '_test.go:[0-9]+:' /tmp/p58-mutant-$name.log | head -2 | cut -c1-300
    fi
  fi
  cp /tmp/p58-original "$file"
}
case "${1:-focused}" in
focused)
  step gofmt bash -c '[ -z "$(gofmt -l internal/lower/)" ]'
  step vet go vet ./internal/lower/
  step pins go test -count=1 -v -run '^TestPlaceholderSlots' ./internal/lower/
  grep -E '^--- (PASS|FAIL): TestPlaceholderSlots[A-Za-z]*(_[0-9])? ' /tmp/p58-pins.log
  grep -E 'programs took a cached answer|programs fall in exactly one' /tmp/p58-pins.log | sed 's/^ *//' | cut -c1-200
  mutant coverage internal/lower/placeholder_cache_test.go 1 '^TestPlaceholderSlotsShardsCoverEveryProgram$' 'TestPlaceholderSlotsShardsCoverEveryProgram' \
    's/^\treturn selected$/\tif shard == 3 { selected = selected[:len(selected)-1] }; return selected/'
  mutant empty internal/lower/placeholder.go 4 '^TestPlaceholderSlotsCachedAgreeWithAWalk_[0-3]$' 'TestPlaceholderSlotsCachedAgreeWithAWalk_[0-3]' \
    's/sources: maps.Clone(l.result.PlaceholderSources), views/sources: ir.SortedMap[string, string]{}, views/'
  ;;
whole)
  step lower go test -count=1 -timeout 40m ./internal/lower/
  grep -E '^(ok|FAIL|---)' /tmp/p58-lower.log | head -8
  ;;
esac
git status --porcelain | grep -v '^??' | head -3
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
