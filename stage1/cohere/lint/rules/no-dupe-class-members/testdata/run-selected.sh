#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
classmembers_selected_dir=$(mktemp -d)
trap 'rm -rf "$classmembers_selected_dir"' EXIT
python3 - "$root" "$classmembers_selected_dir/overlay.json" <<'PY'
import json,sys
from pathlib import Path
root=Path(sys.argv[1])
json.dump({'Replace':{str(root/'stage1/cohere/lint/adamic_classmembers_selected_test.go'):str(root/'stage1/cohere/lint/rules/no-dupe-class-members/testdata/selected_test.go')}},open(sys.argv[2],'w'))
PY
cd "$root"
go run ./cmd/lint-registry >/dev/null
go test -overlay="$classmembers_selected_dir/overlay.json" ./stage1/cohere/lint -run '^TestNoDupeClassMembersSelected$|^TestMutants$/^no-dupe-class-members-static-flag-lost$' -count=1 -v -timeout=25m
