#!/usr/bin/env bash
set -euo pipefail
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
repository=/tmp/void-meter-worktree
scratch=/tmp/void-meter
logs=$scratch/logs
mkdir -p "$logs"
cd "$repository"
# Reuse the identical, setup-verified gitlink without changing tracked sources.
rmdir cohere
ln -s /workspace/adamic/cohere cohere
build_census() {
    local source=$1 destination=$2 logs=$3
    mkdir -p "$destination"
    (cd "$source" && go build -o "$destination/census" ./stage3/census/tool) > "$logs/build.log" 2>&1
    python3 "$repository/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$logs/latent-overlay.log" 2>&1
    gofmt -w "$destination/latent-overlay/"*.go
    python3 "$repository/stage3/meter/entry_overlay.py" "$destination/latent-overlay" "$destination/entry-overlay" > "$logs/entry-overlay.log" 2>&1
    gofmt -w "$destination/entry-overlay/"*.go
    (cd "$source" && go build -buildvcs=false -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool) > "$logs/entry-build.log" 2>&1
    (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$logs/latent-build.log" 2>&1
}
build_census "$repository" "$scratch" "$logs"
upstream_url=$(python3 -c 'import json; print(json.load(open("stage3/source.json"))["repository"])')
upstream_tag=$(python3 -c 'import json; print(json.load(open("stage3/source.json"))["tag"])')
git clone --depth 1 --branch "$upstream_tag" "$upstream_url" "$scratch/upstream" > "$logs/upstream.log" 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=$scratch/upstream
if [[ -f stage3/api/package-lock.json ]]; then
    (cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund) > "$logs/packages.log" 2>&1
fi
bash stage3/apply.sh "$scratch/adapted" > "$logs/apply.log" 2>&1
start=$(date +%s)
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$scratch/latent-census" "$scratch/adapted/src/compiler" "$scratch/latent.jsonl" > "$logs/latent.log" 2>&1
end=$(date +%s)
printf '%s\n' "$((end-start))" > "$scratch/latent-seconds.txt"
