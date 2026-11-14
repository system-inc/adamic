# #qc5jp57's proof: buildcache and compilercensus whole, then five mutants, each one compiling and turning its pin red.
# The head is the first commit under this one not titled proof:. Never merged.
set -u
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/qc-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; exit 3; }
echo "head $(git log --format=%H -1 --invert-grep --grep='^proof: ' HEAD) on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > /tmp/qc-$name.log 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' /tmp/qc-$name.log | head -8 | cut -c1-240; bad=1; }; }
new='^(TestADependencysMarkdownFileMovesNoKey|TestADependencysGoSourceFilesMoveItsKey|TestAnEmbeddingDependencysMarkdownFileMovesItsKey|TestThePackageUnderTestsMarkdownFileMovesItsKey|TestADependencysSubdirectoryMovesNoKey|TestACgoDependencysIncludedSubdirectoryMovesItsKey|TestNoTrackedIncludeNamesAFileAPackageListingLeavesOut|TestARecipeListsADependencysDirectoryWithoutItsTestFiles|TestATracedListingCountsTestFiles)$'
case "${1:-focused}" in
focused)
  step gofmt bash -c '[ -z "$(gofmt -l internal/buildcache/)" ]'
  step vet go vet ./internal/buildcache/
  step pins go test -count=1 -v -run "$new" ./internal/buildcache/
  grep -E '^(--- (PASS|FAIL|SKIP)|ok|FAIL)' /tmp/qc-pins.log | head -12
  grep -E 'C, header and Go files|cgo' /tmp/qc-pins.log | head -2 | cut -c1-200
  mutant() { local label=$1 file=$2 old=$3 new=$4 test=$5
    python3 -I - "$file" "$old" "$new" <<'PY' || { echo "FAIL mutant $label didn't apply"; bad=1; return; }
import sys
path, old, new = sys.argv[1:]
text = open(path).read()
if text.count(old) != 1: sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
    go test -count=1 -v -run "^$test\$" ./internal/buildcache/ > /tmp/qc-m-$label.log 2>&1
    if grep -q -E 'build failed|\[setup failed\]|declared and not used' /tmp/qc-m-$label.log; then echo "FAIL mutant $label didn't compile"; grep -E '\.go:[0-9]+:' /tmp/qc-m-$label.log | head -2; bad=1
    elif grep -q -E "^--- FAIL: $test " /tmp/qc-m-$label.log; then echo "PASS mutant $label red: $test"
    elif grep -q -E "^--- SKIP: $test " /tmp/qc-m-$label.log; then echo "FAIL mutant $label: $test skipped"; grep -m1 -A1 -E "^--- SKIP" /tmp/qc-m-$label.log | tail -1 | cut -c1-200; bad=1
    else echo "FAIL mutant $label survived: $test"; bad=1; fi
    git checkout -q -- "$file"; }
  mutant every-name internal/buildcache/reads.go 'return name == "go.mod" || slices.Contains(packageSourceExtensions, filepath.Ext(name)) && !strings.HasSuffix(name, "_test.go")' 'return !strings.HasSuffix(name, "_test.go")' TestADependencysMarkdownFileMovesNoKey
  mutant embed-narrowed internal/buildcache/settle.go '|| len(listed.EmbedPatterns) > 0 ' '' TestAnEmbeddingDependencysMarkdownFileMovesItsKey
  mutant go-mod-dropped internal/buildcache/reads.go 'return name == "go.mod" || ' 'return ' TestADependencysGoSourceFilesMoveItsKey
  mutant only-dash-i internal/buildcache/settle.go 'var includePathOptions = []string{"-isystem", "-iquote", "-idirafter", "-I"}' 'var includePathOptions = []string{"-I"}' TestACgoDependencysIncludedSubdirectoryMovesItsKey
  mutant h-dropped internal/buildcache/reads.go '".cxx", ".h", ' '".cxx", ' TestNoTrackedIncludeNamesAFileAPackageListingLeavesOut
  ;;
whole)
  step buildcache go test -count=1 -timeout 40m ./internal/buildcache/...
  grep -E '^(ok|FAIL)' /tmp/qc-buildcache.log | head -6
  step compilercensus go test -count=1 -timeout 20m ./internal/compilercensus/...
  ;;
esac
git status --porcelain | grep -v '^??' | head -3
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
