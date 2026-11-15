# #3z7bv9q's proof, sealed: bash proof/performance/3z7bv9q.sh focused|measure, after environment.sh and a job-local
# git commit (mutants restore with git checkout). Fifteen mutants, each compiling and turning its pin red. Never merged.
set -u
echo "on $(hostname), $(nproc) cores, euid $(id -u), load $(cut -d' ' -f1-3 /proc/loadavg)"
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/tz-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/tz-$name.log" | head -8 | cut -c1-240; bad=1; }; }
case "${1:-focused}" in
focused)
  step gofmt bash -c '[ -z "$(gofmt -l internal/buildcache/)" ]'
  step vet go vet ./internal/buildcache/ ./internal/buildcache/cmd/traced/
  step pins go test -count=1 -v -run 'Unpacked|NoWriteProof|ReadOnlyMount|ManifestTakesGits|Remount|Link|Submodule|Kernel' ./internal/buildcache/
  grep -E '^(--- (FAIL|SKIP)|ok|FAIL)' "$TMPDIR/tz-pins.log" | head -10
  step traced go test -count=1 -v -run 'TestTheReportsFileNeverReachesTheCommand|TestNeverStartedIsFourOnlyOnceTheReportSaysSo' ./internal/buildcache/cmd/traced/
  mutant() { local label=$1 file=$2 old=$3 new=$4 test=$5 package=$6
    python3 -I - "$file" "$old" "$new" <<'PY' || { echo "FAIL mutant $label didn't apply"; bad=1; return; }
import sys
path, old, new = sys.argv[1:]
text = open(path).read()
if text.count(old) != 1: sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
    go test -count=1 -v -run "^$test\$" "$package" > "$TMPDIR/tz-m-$label.log" 2>&1
    if grep -q -E 'build failed|\[setup failed\]|declared and not used|imported and not used' "$TMPDIR/tz-m-$label.log"; then echo "FAIL mutant $label didn't compile"; grep -E '\.go:[0-9]+:' "$TMPDIR/tz-m-$label.log" | head -2; bad=1
    elif grep -q -E "^--- FAIL: $test " "$TMPDIR/tz-m-$label.log"; then echo "PASS mutant $label red: $test"
    elif grep -q -E "^--- SKIP: $test " "$TMPDIR/tz-m-$label.log"; then echo "FAIL mutant $label: $test skipped"; bad=1
    else echo "FAIL mutant $label survived: $test"; bad=1; fi
    git checkout -q -- "$file"; }
  u=internal/buildcache/unpacked_state.go; c=internal/buildcache/closed_linux.go; m=internal/buildcache/cmd/traced/main.go; b=./internal/buildcache/; t=./internal/buildcache/cmd/traced/
  w=TestTheNoWriteProofRefusesWhatTheProcessCouldChange; o=TestAReadOnlyMountHoldsOnlyWhenEveryMountUnderTheRootIsReadOnly
  mutant owner-unchecked $u 'if owner == uid {' 'if false && owner == uid {' $w $b
  mutant files-unchecked $u 'if !entry.IsDir() && !entry.Type().IsRegular() {' 'if !entry.IsDir() {' $w $b
  mutant caps-unchecked $u 'if capabilities&(1<<capability.bit) != 0 {' 'if false && capabilities&(1<<capability.bit) != 0 {' $w $b
  mutant capprm-unread $u 'return effective | permitted, nil' 'return effective, nil' $w $b
  mutant ids-unread $u 'if id != ids[0] {' 'if false && id != ids[0] {' $w $b
  mutant mount-ignores-sysadmin $u 'if capabilities&(1<<sysAdminCapability) != 0 {' 'if false && capabilities&(1<<sysAdminCapability) != 0 {' TestTheReadOnlyMountProofRefusesAProcessThatCanRemount $b
  mutant proof-skipped $u 'if held, err = proof(root); err != nil {' 'if held, err = "skipped", error(nil); err != nil {' TestAWritableUnpackedSourceIsRefused $b
  mutant mount-string-prefix $u 'if under(mount.point, root) {' 'if strings.HasPrefix(mount.point, root) {' $o $b
  mutant mount-device-skipped $u 'if mount.device == mounts[own].device && (under(mount.filesystemRoot, rootInFilesystem)' 'if false && (under(mount.filesystemRoot, rootInFilesystem)' $o $b
  mutant superblock-field $u 'options: fields[5]})' 'options: fields[len(fields)-1]})' $o $b
  mutant link-root-unresolved $u 'uid, err := noWriteProof(resolved, reads)' 'uid, err := noWriteProof(root, reads)' TestARootThatIsALinkIsProvenOverItsTree $b
  mutant submodule-head-skipped $u 'if _, ok := heads[repository]; !ok {' 'if _, ok := heads[repository]; false && !ok {' TestAManifestMissingASubmodulesHeadIsRefused $b
  mutant kernel-never-writable $c 'case err == nil:' 'case false:' TestTheKernelAnswersWhatThisProcessMayWrite $b
  mutant report-left-set $m 'os.Unsetenv("ADAMIC_TRACE_REPORT")' '' TestTheReportsFileNeverReachesTheCommand $t
  mutant write-error-ignored $m 'writeError != nil {' 'false && writeError != nil {' TestNeverStartedIsFourOnlyOnceTheReportSaysSo $t
  ;;
measure)
  ADAMIC_MEASURE_CLOSURE_WALK=$PWD go test -count=1 -run '^TestMeasureTheClosureWalk$' -v ./internal/buildcache/ > "$TMPDIR/tz-measure.log" 2>&1; code=$?
  grep -E '_test.go:[0-9]+:|^(--- |ok|FAIL)' "$TMPDIR/tz-measure.log" | head -8 | cut -c1-240
  [ $code = 0 ] || bad=1
  ;;
esac
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
