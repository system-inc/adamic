#!/usr/bin/env bash
set -euo pipefail
hir_rule_root="$(git rev-parse --show-toplevel)"
hir_rule_overlay="$(mktemp)"
trap 'rm -f "$hir_rule_overlay"' EXIT
python3 - "$hir_rule_root" "$hir_rule_overlay" <<'PY'
import json, pathlib, sys
root = pathlib.Path(sys.argv[1])
with open(sys.argv[2], 'w') as out:
    json.dump({'Replace': {
        str(root / 'stage1/cohere/lint/stage1_hir_certificate_test.go'):
        str(root / 'stage1/cohere/lint/rules/react-hooks-static-components/testdata/certification_test.go')
    }}, out)
PY
cd "$hir_rule_root"
go test -count=1 -v -timeout=3h -overlay "$hir_rule_overlay" ./stage1/cohere/lint -run TestHIRStaticComponentsCertificate
