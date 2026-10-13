#!/usr/bin/env bash
set -euo pipefail
# Supply the adapted tree produced by stage3/apply.sh at 9b8ebd77 and that
# revision's language-decision/results.json. Every pinned source hash is verified.
repository=$(cd "$(dirname "$0")/../.." && pwd)
measurement=$(mktemp -d)
trap 'rm -rf "$measurement"' EXIT
python3 "$repository/stage3/census/latent/make_overlay.py" "$repository" "$measurement" > "$measurement/setup.log" 2>&1
python3 - "$repository" "$measurement/overlay.json" <<'PY'
import json, sys
root, path = sys.argv[1:]
overlay = json.load(open(path))
overlay['Replace'][root + '/internal/lower/checked_writes_census_test.go'] = root + '/stage3/checked-writes/census_test.go.txt'
with open(path, 'w') as output:
    json.dump(overlay, output)
PY
cd "$repository"
CHECKED_WRITES_TREE="$1" CHECKED_WRITES_RECORDS="$2" CHECKED_WRITES_OUTPUT="$repository/stage3/checked-writes/census.json" go test -overlay "$measurement/overlay.json" ./internal/lower -run '^TestCheckedWritesCensus$' -count=1 -v -timeout 10m > "${3:-/tmp/checked-wider-census.log}" 2>&1
