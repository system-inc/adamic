#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
if [[ -f /workspace/adamic-tools/env.sh ]]; then source /workspace/adamic-tools/env.sh; fi
mutant_id=${1:?provide M1 through M12, or P1 through P3}
case "$mutant_id" in M[1-9]|M1[0-2]|P[1-3]) ;; *) exit 2;; esac
base=review/test-audit/internal-regexp-parser
diff_path="$base/diffs/$mutant_id.diff"
git apply --check "$diff_path"
git apply "$diff_path"
trap 'git apply -R "$diff_path"' EXIT
if [[ "$mutant_id" =~ ^M1[0-2]$ || "$mutant_id" =~ ^P[23]$ ]]; then go vet ./internal/unicodeproperties/; else go vet ./internal/regexp/; fi
ADAMIC_BUILD_CACHE_DIR="/tmp/u074/replay-cache/$mutant_id" timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run . > "$base/replay-$mutant_id.log" 2>&1
