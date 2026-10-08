#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
react_selected_dir=$(mktemp -d)
trap 'rm -rf "$react_selected_dir"' EXIT
python3 - "$root" "$react_selected_dir/overlay.json" <<'PY'
import json,sys
from pathlib import Path
root=Path(sys.argv[1])
json.dump({'Replace':{str(root/'stage1/cohere/lint/adamic_react_selected_test.go'):str(root/'stage1/cohere/lint/rules/react-state-in-constructor/testdata/selected_test.go')}},open(sys.argv[2],'w'))
PY
cd "$root"
go run ./cmd/lint-registry >/dev/null
go test -overlay="$react_selected_dir/overlay.json" ./stage1/cohere/lint -run '^TestReactStateInConstructorSelected$|^TestMutants$/^accept-plain-classes$' -count=1 -v -timeout=20m
