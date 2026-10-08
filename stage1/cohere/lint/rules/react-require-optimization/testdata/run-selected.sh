#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
optimization_selected_dir=$(mktemp -d)
trap 'rm -rf "$optimization_selected_dir"' EXIT
python3 - "$root" "$optimization_selected_dir/overlay.json" <<'PY'
import json,sys
from pathlib import Path
root=Path(sys.argv[1])
json.dump({'Replace':{str(root/'stage1/cohere/lint/adamic_optimization_selected_test.go'):str(root/'stage1/cohere/lint/rules/react-require-optimization/testdata/selected_test.go')}},open(sys.argv[2],'w'))
PY
cd "$root"
go run ./cmd/lint-registry >/dev/null
go test -overlay="$optimization_selected_dir/overlay.json" ./stage1/cohere/lint -run '^TestReactRequireOptimizationSelected$|^TestMutants$/^require-optimization-accept-class-field$' -count=1 -v -timeout=25m
