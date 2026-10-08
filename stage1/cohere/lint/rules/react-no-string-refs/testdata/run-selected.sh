#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
refs_selected_dir=$(mktemp -d)
trap 'rm -f "$refs_selected_dir/overlay.json"; rmdir "$refs_selected_dir"' EXIT
python3 - "$root" "$refs_selected_dir/overlay.json" <<'PY'
import json,sys
from pathlib import Path
root=Path(sys.argv[1])
json.dump({'Replace':{str(root/'stage1/cohere/lint/adamic_string_refs_selected_test.go'):str(root/'stage1/cohere/lint/rules/react-no-string-refs/testdata/selected_test.go')}},open(sys.argv[2],'w'))
PY
cd "$root"
go run ./cmd/lint-registry
go test -overlay="$refs_selected_dir/overlay.json" ./stage1/cohere/lint -run '^TestReactNoStringRefsSelected$|^TestMutants$/^string-refs-accept-string-attribute$' -count=1 -v -timeout=30m
