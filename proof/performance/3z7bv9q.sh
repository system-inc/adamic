# #3z7bv9q's proof: pins, eight mutants each compiling and turning its pin red, then buildcache, traced and
# compilercensus whole. The head is the first commit under this one not titled proof:. Never merged.
set -u
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/tz-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; exit 3; }
echo "head $(git log --format=%H -1 --invert-grep --grep='^proof: ' HEAD) on $(hostname), $(nproc) cores, euid $(id -u), load $(cut -d' ' -f1-3 /proc/loadavg)"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > /tmp/tz-$name.log 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' /tmp/tz-$name.log | head -8 | cut -c1-240; bad=1; }; }
case "${1:-focused}" in
focused)
  step gofmt bash -c '[ -z "$(gofmt -l internal/buildcache/)" ]'
  step vet go vet ./internal/buildcache/ ./internal/buildcache/cmd/traced/
  step pins go test -count=1 -v -run 'Unpacked|NoWriteProof|ReadOnlyMount|ManifestTakesGits|Remount' ./internal/buildcache/
  grep -E '^(--- (PASS|FAIL|SKIP)|ok|FAIL)' /tmp/tz-pins.log | head -14
  step traced go test -count=1 -v -run 'TestTheReportsFileNeverReachesTheCommand' ./internal/buildcache/cmd/traced/
  mutant() { local label=$1 file=$2 old=$3 new=$4 test=$5 package=$6
    python3 -I - "$file" "$old" "$new" <<'PY' || { echo "FAIL mutant $label didn't apply"; bad=1; return; }
import sys
path, old, new = sys.argv[1:]
text = open(path).read()
if text.count(old) != 1: sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
    go test -count=1 -v -run "^$test\$" "$package" > /tmp/tz-m-$label.log 2>&1
    if grep -q -E 'build failed|\[setup failed\]|declared and not used|imported and not used' /tmp/tz-m-$label.log; then echo "FAIL mutant $label didn't compile"; grep -E '\.go:[0-9]+:' /tmp/tz-m-$label.log | head -2; bad=1
    elif grep -q -E "^--- FAIL: $test " /tmp/tz-m-$label.log; then echo "PASS mutant $label red: $test"
    elif grep -q -E "^--- SKIP: $test " /tmp/tz-m-$label.log; then echo "FAIL mutant $label: $test skipped"; bad=1
    else echo "FAIL mutant $label survived: $test"; bad=1; fi
    git checkout -q -- "$file"; }
  u=internal/buildcache/unpacked_state.go; b=./internal/buildcache/
  mutant owner-unchecked $u 'if owner == reads.effectiveUid {' 'if false && owner == reads.effectiveUid {' TestTheNoWriteProofRefusesWhatTheProcessCouldChange $b
  mutant files-unchecked $u 'if !entry.IsDir() && !entry.Type().IsRegular() {' 'if !entry.IsDir() {' TestTheNoWriteProofRefusesWhatTheProcessCouldChange $b
  mutant caps-unchecked $u 'if capabilities&(1<<capability.bit) != 0 {' 'if false && capabilities&(1<<capability.bit) != 0 {' TestTheNoWriteProofRefusesWhatTheProcessCouldChange $b
  mutant mount-ignores-sysadmin $u 'if capabilities&(1<<sysAdminCapability) != 0 {' 'if false && capabilities&(1<<sysAdminCapability) != 0 {' TestTheReadOnlyMountProofRefusesAProcessThatCanRemount $b
  mutant proof-skipped $u 'if held, err = proof(root); err != nil {' 'if held, err = "skipped", error(nil); err != nil {' TestAWritableUnpackedSourceIsRefused $b
  mutant mount-device-skipped $u 'if mount.device == mounts[own].device && (under(mount.filesystemRoot, rootInFilesystem)' 'if false && (under(mount.filesystemRoot, rootInFilesystem)' TestAReadOnlyMountHoldsOnlyWhenEveryMountUnderTheRootIsReadOnly $b
  mutant mount-string-prefix $u 'if under(point, root) && ' 'if strings.HasPrefix(point, root) && ' TestAReadOnlyMountHoldsOnlyWhenEveryMountUnderTheRootIsReadOnly $b
  mutant report-left-set internal/buildcache/cmd/traced/main.go 'os.Unsetenv("ADAMIC_TRACE_REPORT")' '' TestTheReportsFileNeverReachesTheCommand ./internal/buildcache/cmd/traced/
  ;;
whole)
  step buildcache go test -count=1 -timeout 40m ./internal/buildcache/...
  grep -E '^(ok|FAIL)' /tmp/tz-buildcache.log | head -8
  step compilercensus go test -count=1 -timeout 20m ./internal/compilercensus/...
  ;;
esac
git status --porcelain | grep -v '^??' | head -3
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
