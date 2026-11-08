#!/usr/bin/env bash
set -u
cd /workspace/adamic-rule-g-2594029 || exit 1
source /workspace/adamic-tools/env.sh
proof=/workspace/adamic/evidence/e5531c8
pin=e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1
restore() { git restore --source="$pin" -- internal/lower/invariance.go; }
trap restore EXIT
run() {
  local label=$1
  shift
  printf '%q ' "$@" > "$proof/$label.command"
  printf '\n' >> "$proof/$label.command"
  "$@" > "$proof/$label.log" 2>&1
  printf '%s\n' "$?" > "$proof/$label.exit"
}
run 10-baseline go test -count=1 -run 'TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet' -v ./internal/lower/
python3 - <<'PY'
from pathlib import Path
p=Path('internal/lower/invariance.go'); s=p.read_text()
a='if l.signatureAny(instantiated, target) != nil && l.signatureAny(fromSignatures[0], target) == nil {'
assert s.count(a)==1
p.write_text(s.replace(a,'if false {'))
PY
if [ "$?" -ne 0 ]; then exit 1; fi
git diff -- internal/lower/invariance.go > "$proof/20-mutant-b.patch"
run 21-mutant-b go test -count=1 -run 'TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet' -v ./internal/lower/
restore
run 22-restoration git diff --exit-code "$pin" --
run 23-final-status git status --short
cat "$proof/10-baseline.log" "$proof/21-mutant-b.log"
