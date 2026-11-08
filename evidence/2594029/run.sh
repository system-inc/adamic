#!/usr/bin/env bash
set -u
cd /workspace/adamic-rule-g-2594029 || exit 1
proof=/workspace/adamic/evidence/2594029
pin=2594029da13920577f63b2a4798325d7d76d8f4f
while [ ! -f "$proof/06-submodules-retry.exit" ]; do sleep 2; done
if [ "$(cat "$proof/06-submodules-retry.exit")" != 0 ]; then exit 1; fi
source /workspace/adamic-tools/env.sh
run() {
  local label=$1
  shift
  printf '%q ' "$@" > "$proof/$label.command"
  printf '\n' >> "$proof/$label.command"
  "$@" > "$proof/$label.log" 2>&1
  local result=$?
  printf '%s\n' "$result" > "$proof/$label.exit"
  return 0
}
restore() {
  git restore --source="$pin" -- internal/lower/invariance.go internal/lower/rule_g_test.go
}
trap restore EXIT
run 07-go-version go version
run 08-clang-version clang --version
run 09-node-version node --version
python3 - <<'PY'
from pathlib import Path
p=Path('internal/lower/rule_g_test.go')
s=p.read_text()
a='\t\t\t_, err := lowerSourceAs(t, source, extension)'
assert s.count(a)==1
p.write_text(s.replace(a,a+'\n\t\t\tt.Logf("stops at: %v", err)'))
PY
if [ "$?" -ne 0 ]; then exit 1; fi
git diff -- internal/lower/rule_g_test.go > "$proof/10-temporary-timer-log.patch"
run 11-focused go test -count=1 -run 'TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod' -v ./internal/lower/
restore
python3 - <<'PY'
from pathlib import Path
p=Path('internal/lower/invariance.go')
s=p.read_text()
a='!slices.Equal(fromSignatures[0].TypeParameters(), toSignatures[0].TypeParameters())'
assert s.count(a)==1
p.write_text(s.replace(a,'len(toSignatures[0].TypeParameters()) == 0'))
PY
if [ "$?" -ne 0 ]; then exit 1; fi
git diff -- internal/lower/invariance.go > "$proof/20-mutant-a.patch"
run 21-mutant-a go test -count=1 -run 'TestASetTimeout|TestAGenericMethodWithANarrowerConstraint' -v ./internal/lower/
restore
run 22-restored-after-a git diff --exit-code "$pin" -- internal/lower/
python3 - <<'PY'
from pathlib import Path
p=Path('internal/lower/invariance.go')
s=p.read_text()
a='if l.signatureAny(instantiated, target) != nil && l.signatureAny(fromSignatures[0], target) == nil {'
assert s.count(a)==1
p.write_text(s.replace(a,'if false {'))
PY
if [ "$?" -ne 0 ]; then exit 1; fi
git diff -- internal/lower/invariance.go > "$proof/30-mutant-b.patch"
run 31-mutant-b go test -count=1 -run 'TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet' -v ./internal/lower/
restore
run 32-restored-after-b git diff --exit-code "$pin" --
run 33-final-status git status --short
printf 'Steps 2–4 finished.\n'
