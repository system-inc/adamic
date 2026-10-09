#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
direct_state_overlay=$(mktemp /tmp/direct-state-overlay.XXXXXX)
python3 - "$root" "$direct_state_overlay" <<'OVERLAY'
import json, sys
from pathlib import Path
root = Path(sys.argv[1])
with open(sys.argv[2], 'w') as output:
    json.dump({'Replace': {str(root / 'stage1/cohere/lint/direct_state_selected_test.go'): str(root / 'stage1/cohere/lint/rules/react-no-direct-mutation-state/testdata/selected_test.go')}}, output)
OVERLAY
cd "$root"
go run ./cmd/lint-registry
go test -overlay="$direct_state_overlay" ./stage1/cohere/lint -run '^TestReactNoDirectMutationStateSelected$|^TestMutants$/^direct-state-constructor-call-lost$' -count=1 -v -timeout=30m
