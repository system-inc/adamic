#!/usr/bin/env bash
set -euo pipefail
lifecycle_root=$(git rev-parse --show-toplevel)
lifecycle_logs=$(mktemp -d "${TMPDIR:-/tmp}/react-lifecycle-selected.XXXXXX")
python3 - "$lifecycle_root" "$lifecycle_logs/overlay.json" <<'PY'
import json, sys
from pathlib import Path
root = Path(sys.argv[1])
Path(sys.argv[2]).write_text(json.dumps({'Replace': {
    str(root / 'stage1/cohere/lint/react_no_arrow_function_lifecycle_selected_test.go'):
    str(root / 'stage1/cohere/lint/rules/react-no-arrow-function-lifecycle/testdata/selected_test.go')
}}))
PY
cd "$lifecycle_root"
go run ./cmd/lint-registry > "$lifecycle_logs/registry.log" 2>&1
echo "Logs: $lifecycle_logs"
go test -overlay="$lifecycle_logs/overlay.json" ./stage1/cohere/lint -run 'TestReactNoArrowFunctionLifecycle|TestMutants/react_lifecycle_static_list_omitted' -count=1 -v -timeout=30m > "$lifecycle_logs/selected.log" 2>&1
