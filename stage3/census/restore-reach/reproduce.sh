#!/usr/bin/env bash
set -euo pipefail
# Run in a private detached worktree of 949a2198, with setup already complete.
# All caches and the adapted tree must be outside the shared checkout.
: "${RESTORE_REACH_TREE:?set to the completed adapted tree}"
: "${TMPDIR:?set to scratch}"
: "${GOTMPDIR:?set to scratch}"
: "${GOCACHE:?set to scratch}"
probe=internal/lower/restore_reach_probe_test.go
if [[ -e $probe ]]; then echo "refusing to overwrite $probe" >&2; exit 1; fi
trap 'rm -f "$probe"' EXIT
cp stage3/census/restore-reach/probe_test.go.txt "$probe"
node stage3/census/restore-reach/inventory.cjs "$RESTORE_REACH_TREE"
go test ./internal/lower -run '^TestRestoreReachProbe$' -v -count=1
