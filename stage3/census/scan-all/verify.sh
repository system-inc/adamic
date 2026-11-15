#!/usr/bin/env bash
set -euo pipefail
repo=$(git rev-parse --show-toplevel)
cd "$repo"
: "${SCAN_ALL_SCRATCH:?set a scratch directory outside the checkout}"
mkdir -p "$SCAN_ALL_SCRATCH"
python3 stage3/census/scan-all/make_overlay.py "$SCAN_ALL_SCRATCH/overlay"
go build -overlay "$SCAN_ALL_SCRATCH/overlay/overlay.json" -o "$SCAN_ALL_SCRATCH/scan-all" ./stage3/census/scan-all/tool/main.go
python3 - "$repo" "$SCAN_ALL_SCRATCH" <<'PYTHON'
import json,pathlib,sys
repo,scratch=map(pathlib.Path,sys.argv[1:])
(scratch/'production-overlay.json').write_text(json.dumps({'Replace':{str(repo/'stage3/census/scan-all/production/main.go'):str(repo/'stage3/census/scan-all/production.go.txt')}}))
PYTHON
go build -overlay "$SCAN_ALL_SCRATCH/production-overlay.json" -o "$SCAN_ALL_SCRATCH/production" ./stage3/census/scan-all/production/main.go
export SCAN_ALL_BINARY="$SCAN_ALL_SCRATCH/scan-all" SCAN_ALL_PRODUCTION="$SCAN_ALL_SCRATCH/production"
go test -v -count=1 -parallel=2 -timeout=60s ./stage3/census/scan-all
for name in CountStop CountClean RestStop RestClean RecordStop RecordClean CycleStop CycleClean; do
  if SCAN_ALL_MUTANT_FIRST_ERROR=1 go test -count=1 -timeout=60s ./stage3/census/scan-all -run "^TestScanAll${name}AfterEarlierScanStop$" > "$SCAN_ALL_SCRATCH/mutant-$name.log" 2>&1; then
    echo "first-error mutant survived: $name" >&2; exit 1
  fi
  grep -Eq "continuation-target|continuation-completion" "$SCAN_ALL_SCRATCH/mutant-$name.log"
  grep -- "--- FAIL: TestScanAll${name}AfterEarlierScanStop" "$SCAN_ALL_SCRATCH/mutant-$name.log"
done
