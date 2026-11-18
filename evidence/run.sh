#!/usr/bin/env bash
set -euo pipefail
task=$1
export TMPDIR="$task/tmp"
export CENSUS_MODULE_FORMAT=CommonJSBundle
unset ADAMIC_PROGRAM_REGION ADAMIC_OWNERSHIP_QUERY
build() (
    label=$1
    dest="$task/$label-bin"
    log="$task/evidence/$label"
    cd "$task/$label"
    mkdir -p "$dest"
    git rev-parse HEAD > "$log/compiler.txt"
    go version >> "$log/compiler.txt"
    git status --porcelain --untracked-files=no >> "$log/compiler.txt"
    go build -o "$dest/census" ./stage3/census/tool > "$log/build.log" 2>&1
    python3 stage3/census/latent/make_overlay.py "$PWD" "$dest/latent-overlay" > "$log/latent-overlay.log" 2>&1
    gofmt -w "$dest/latent-overlay/"*.go
    python3 stage3/meter/entry_overlay.py "$dest/latent-overlay" "$dest/entry-overlay" > "$log/entry-overlay.log" 2>&1
    gofmt -w "$dest/entry-overlay/"*.go
    go build -buildvcs=false -overlay="$dest/entry-overlay/overlay.json" -o "$dest/entry-census" ./stage3/census/latent/tool > "$log/entry-build.log" 2>&1
    (cd "$dest" && sha256sum census entry-census) > "$log/binaries.sha256"
)
prepare() (
    cd "$task/base"
    export ADAMIC_TYPESCRIPT_SOURCE="$task/upstream"
    git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD > "$task/evidence/upstream.txt"
    test "$(cat "$task/evidence/upstream.txt")" = 050880ce59e30b356b686bd3144efe24f875ebc8
    node --version >> "$task/evidence/upstream.txt"
    (cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund) > "$task/evidence/packages.log" 2>&1
    bash stage3/apply.sh "$task/adapted" > "$task/evidence/apply.log" 2>&1
    (cd "$task/adapted" && find src -type f -print0 | sort -z | xargs -0 sha256sum) > "$task/evidence/adapted-src.sha256"
)
measure() (
    label=$1
    dest="$task/$label-bin"
    log="$task/evidence/$label"
    cd "$task/$label"
    "$dest/census" "$task/adapted/src/compiler" "$log/census.jsonl" > "$log/census.log" 2>&1 &
    census_pid=$!
    LATENT_TIMES="$log/latent-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 "$dest/entry-census" "$task/adapted/src/tsc/tsc.ts" "$log/latent.jsonl" > "$log/latent.log" 2>&1 &
    latent_pid=$!
    census_status=0
    latent_status=0
    wait "$census_pid" || census_status=$?
    wait "$latent_pid" || latent_status=$?
    printf 'census: %s\nentry-census: %s\n' "$census_status" "$latent_status" > "$log/exit-status.txt"
    (( census_status == 0 && latent_status == 0 ))
)
case $2 in
    build) build "$3";;
    prepare) prepare;;
    measure) measure "$3";;
esac
