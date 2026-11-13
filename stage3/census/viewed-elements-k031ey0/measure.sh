#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
mkdir -p "$TMPDIR"
for side in main head; do
    source_tree=/workspace/k031ey0/$side
    destination=/workspace/k031ey0/$side-binaries
    mkdir -p "$destination"
    cd "$source_tree"
    python3 stage3/census/latent/make_overlay.py "$source_tree" "$destination/latent-overlay" > "/workspace/k031ey0/evidence/$side-latent-overlay.log" 2>&1
    gofmt -w "$destination/latent-overlay/"*.go
    python3 stage3/meter/entry_overlay.py "$destination/latent-overlay" "$destination/entry-overlay" > "/workspace/k031ey0/evidence/$side-entry-overlay.log" 2>&1
    gofmt -w "$destination/entry-overlay/"*.go
    go build -buildvcs=false -p=1 -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool > "/workspace/k031ey0/evidence/$side-entry-build.log" 2>&1
    sha256sum "$destination/entry-census" > "/workspace/k031ey0/evidence/$side-binary.sha256"
    set +e
    LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 ADAMIC_VIEWED_ELEMENTS_LOG=1 LATENT_ONLY_FILE=/workspace/k031ey0/adapted/src/compiler/_namespaces/ts.performance.ts timeout --preserve-status 900 "$destination/entry-census" /workspace/k031ey0/adapted/src/tsc/tsc.ts "/workspace/k031ey0/evidence/$side.jsonl" > "/workspace/k031ey0/evidence/$side.stdout" 2> "/workspace/k031ey0/evidence/$side.stderr"
    child_status=$?
    set -e
    printf '%s\n' "$child_status" > "/workspace/k031ey0/evidence/$side.exit-code"
    printf '%s child exit code: %s\n' "$side" "$child_status"
done
