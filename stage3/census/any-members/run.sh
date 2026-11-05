#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
repository=/workspace/adamic
out=$repository/stage3/census/any-members
scratch=$out/scratch
mkdir -p "$scratch/tmp" "$out/evidence"
export TMPDIR=$scratch/tmp
export GOCACHE=$scratch/go-cache
export GOPATH=$scratch/go-path
export STAGE3_CACHE=$scratch/stage3-cache
export npm_config_cache=$scratch/npm-cache
python3 "$repository/stage3/meter/box.py" "$out/evidence/box.json" start
for label in base head; do
 ref=b0efce6dce1cd7445ed8a570aecac387a4a7d684
 [[ $label == base ]] || ref=66fff4ecc05c3bfb2258cf4bb2a7b0a2b762c59b
 git worktree add --detach "$scratch/$label" "$ref" > "$out/evidence/$label-checkout.log" 2>&1
 git -C "$scratch/$label" -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1 > "$out/evidence/$label-submodules.log" 2>&1
 git -C "$scratch/$label" submodule status --recursive > "$out/evidence/$label-pins.txt"
 git -C "$scratch/$label" status --porcelain --untracked-files=no > "$out/evidence/$label-status.txt"
 if [[ $label == base ]]; then
  git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$scratch/upstream" > "$out/evidence/upstream.log" 2>&1
  export ADAMIC_TYPESCRIPT_SOURCE=$scratch/upstream
  (cd "$scratch/base/stage3/api" && npm ci --ignore-scripts --no-audit --no-fund) > "$out/evidence/packages.log" 2>&1
  bash "$scratch/base/stage3/apply.sh" "$scratch/adapted" > "$out/evidence/apply.log" 2>&1
 fi
 mkdir -p "$scratch/$label-bin"
 python3 "$scratch/base/stage3/census/latent/make_overlay.py" "$scratch/$label" "$scratch/$label-bin/overlay" > "$out/evidence/$label-overlay.log" 2>&1
 gofmt -w "$scratch/$label-bin/overlay/"*.go
 (cd "$scratch/$label" && go build -buildvcs=false -overlay="$scratch/$label-bin/overlay/overlay.json" -o "$scratch/$label-bin/latent-census" ./stage3/census/latent/tool) > "$out/evidence/$label-build.log" 2>&1
 cat /sys/fs/cgroup/memory.events > "$out/evidence/$label-memory-before.txt"
 LATENT_TIMES="$out/evidence/$label-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$scratch/$label-bin/latent-census" "$scratch/adapted/src/compiler" "$out/evidence/$label-latent.jsonl" > "$out/evidence/$label-latent.log" 2>&1 &
 census_pid=$!
 echo "$census_pid" > "$out/evidence/$label-pid.txt"
 date -u +%FT%TZ > "$out/evidence/$label-start.txt"
 set +e
 wait "$census_pid"
 status=$?
 set -e
 echo "$status" > "$out/evidence/$label-exit.txt"
 date -u +%FT%TZ > "$out/evidence/$label-end.txt"
 cat /sys/fs/cgroup/memory.events > "$out/evidence/$label-memory-after.txt"
 [[ $status == 0 ]] || exit "$status"
done
python3 "$repository/stage3/meter/box.py" "$out/evidence/box.json" end
