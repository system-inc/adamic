#!/usr/bin/env bash
set -eu
repo_root=$(git rev-parse --show-toplevel)
mutant_directory=$(mktemp -d)
for mutant_name in flatten header-key flags blanket-warning; do
  test_name=TestUnitSystemHeaderProvenance
  case "$mutant_name" in
    flatten) expected_message="system-header provenance lost" ;;
    header-key) expected_message='output="1\n" want="2\n"' ;;
    flags) test_name=TestUnitCacheFlagsHoldSanitizer; expected_message="sanitized rebuild reused uninstrumented object" ;;
    blanket-warning) expected_message="user-header extension must be rejected: <nil>" ;;
  esac
  python3 - "$repo_root" "$mutant_directory" "$mutant_name" <<'PYTHON'
import json, sys
from pathlib import Path
root, directory, name = map(Path, sys.argv[1:])
mutant = directory / (str(name) + '.go')
mutant.write_bytes((root / 'internal/native/clang_units_evidence' / ('system-header-' + str(name) + '-mutant.go.txt')).read_bytes())
(directory / 'overlay.json').write_text(json.dumps({'Replace': {str(root / 'internal/native/units.go'): str(mutant)}}))
PYTHON
  if go test -overlay="$mutant_directory/overlay.json" ./internal/native -run "^$test_name$" -count=1 > "$mutant_directory/$mutant_name.log" 2>&1; then
    echo "ERROR: $mutant_name survived"
    exit 1
  fi
  cp "$mutant_directory/$mutant_name.log" "/tmp/adamic-header-$mutant_name-reproduced.log"
  if ! rg -F -- "$expected_message" "$mutant_directory/$mutant_name.log" > /dev/null; then
    echo "ERROR: unexpected failure; /tmp/adamic-header-$mutant_name-reproduced.log"
    exit 1
  fi
  if [ "$mutant_name" = flatten ] && ! rg -F -- '-Wnullability-extension' "$mutant_directory/$mutant_name.log" > /dev/null; then
    echo "ERROR: flattening did not expose the nullability extension"
    exit 1
  fi
  echo "$mutant_name caught; /tmp/adamic-header-$mutant_name-reproduced.log"
done
echo "Overlays and logs retained in $mutant_directory"
