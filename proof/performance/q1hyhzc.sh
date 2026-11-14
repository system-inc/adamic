# #q1hyhzc's proof: bash proof/performance/q1hyhzc.sh focused|whole. The parent of this commit is the head.
set -u
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/q1-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; tail -3 /tmp/q1-sub.log; exit 3; }
echo "head $(git log --format=%H -1 --invert-grep --grep='^proof: ' HEAD) on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > /tmp/q1-$name.log 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' /tmp/q1-$name.log | head -12; bad=1; }; return $code; }
pins='^(TestViewSiteIndexIsBuiltOncePerLowering|TestViewSiteAskedAgainWhenItsNameGainsAKey|TestViewRefusesItsFirstSiteInPreorder|TestViewWalkIndexedAgreesWithWhole|TestViewReceiverCallersAreClassified|TestCheckedFieldsAreWrittenThroughMarkCheckedField)$'
case "${1:-focused}" in
focused)
  step gofmt bash -c '[ -z "$(gofmt -l internal/lower/ stage3/census/latent/statecopy/)" ]'
  step vet go vet ./internal/lower/ ./stage3/census/latent/...
  step pins go test -count=1 -v -run "$pins" ./internal/lower/
  grep -E '^(--- (PASS|FAIL)|ok|FAIL)' /tmp/q1-pins.log | head -12
  # Mutant: only names never judged are asked; a name that gains a key isn't asked again.
  sed -i 's/asked = !judged || seen != written\[name\]/asked = !judged; _ = seen/' internal/lower/view_sites.go
  if git diff --quiet internal/lower/view_sites.go; then echo "FAIL mutant didn't apply"; bad=1; else
    go test -count=1 -v -run '^(TestViewSiteAskedAgainWhenItsNameGainsAKey|TestViewWalkIndexedAgreesWithWhole)$' ./internal/lower/ > /tmp/q1-mutant.log 2>&1
    reds=$(grep -c -E '^--- FAIL: (TestViewSiteAskedAgainWhenItsNameGainsAKey|TestViewWalkIndexedAgreesWithWhole) ' /tmp/q1-mutant.log)
    # A mutant that doesn't build kills nothing: J-stnyy86 and J-vsr8hg9 read 0 of 2 because sed left seen unused.
    if grep -q -E 'build failed|declared and not used' /tmp/q1-mutant.log; then echo "FAIL mutant didn't compile"; grep -E '\.go:[0-9]+:' /tmp/q1-mutant.log | head -3; bad=1; fi
    echo "$([ "$reds" = 2 ] && echo PASS || echo FAIL) mutant: $reds of 2 red"; [ "$reds" = 2 ] || bad=1
    grep -E '_test.go:[0-9]+:|view walk comparison' /tmp/q1-mutant.log | head -3 | cut -c1-300
  fi
  git checkout -q -- internal/lower/view_sites.go
  step census go test -count=1 ./stage3/census/latent/...
  step ir go test -count=1 ./internal/ir/
  ;;
whole)
  step lower go test -count=1 -timeout 40m ./internal/lower/
  ;;
fixtures)
  step fixtures go test -count=1 -timeout 40m ./stage3/fixtures/...
  grep -E '^(ok|FAIL|---)' /tmp/q1-fixtures.log | head -12
  ;;
esac
git status --porcelain | grep -v '^??' | head -3
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
