#!/usr/bin/env bash
set -eu
repo_root=$(git rev-parse --show-toplevel)
mutant_directory=$(mktemp -d)
trap 'rm -rf "$mutant_directory"' EXIT
for mutant_name in flags state literal determinism; do
  case "$mutant_name" in
    flags) test_name=TestUnitCacheFlagsHoldSanitizer; expected_message="sanitized rebuild reused uninstrumented object" ;;
    state) test_name=TestUnitsPreserveSharedState; expected_message="shared state (uncached=0)" ;;
    determinism) test_name=TestUnitsPreserveSharedState; expected_message="split changed" ;;
    literal) test_name=TestSplitTokensDoNotRewriteLiterals; expected_message="literal changed" ;;
  esac
  python3 - "$repo_root" "$mutant_directory" "$mutant_name" <<'PYTHON'
import json,sys
from pathlib import Path
root,directory,name=map(Path,sys.argv[1:])
mutant=directory/(str(name)+'.go')
mutant.write_bytes((root/'internal/native/clang_units_evidence'/(str(name)+'-mutant.go.txt')).read_bytes())
(directory/'overlay.json').write_text(json.dumps({'Replace':{str(root/'internal/native/units.go'):str(mutant)}}))
PYTHON
  if go test -overlay="$mutant_directory/overlay.json" ./internal/native -run "^$test_name$" -count=1 > "$mutant_directory/$mutant_name.log" 2>&1; then
    echo "ERROR: $mutant_name survived; log $mutant_directory/$mutant_name.log"
    exit 1
  fi
  if ! rg -F "$expected_message" "$mutant_directory/$mutant_name.log" > /dev/null; then
    echo "ERROR: $mutant_name failed for an unexpected reason; log $mutant_directory/$mutant_name.log"
    exit 1
  fi
  cp "$mutant_directory/$mutant_name.log" "/tmp/adamic-clang-$mutant_name-reproduced.log"
  echo "$mutant_name caught; /tmp/adamic-clang-$mutant_name-reproduced.log"
done
