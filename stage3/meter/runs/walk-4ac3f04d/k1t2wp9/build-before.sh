#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
repository=$PWD
out=$repository/stage3/meter/runs/walk-4ac3f04d/k1t2wp9
scratch=/workspace/adamic-tools/k1t2wp9-before
mkdir -p "$scratch" "$out/before-build"
logs=$out/before-build
source=$scratch/source
destination=$scratch/binaries
git worktree add --detach "$source" 4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288 > "$logs/checkout.log" 2>&1
git -C "$source" -c protocol.file.allow=always -c "url.file://$repository/cohere.insteadOf=https://github.com/system-inc/cohere.git" -c "url.file://$repository/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git" submodule update --init --recursive --depth 1 > "$logs/submodules.log" 2>&1
# Extract the unchanged build_census function from the meter, retaining its exact commands.
python3 - "$repository/stage3/meter/twice-daily.sh" "$scratch/build-function.sh" <<'PY'
from pathlib import Path
import sys
s=Path(sys.argv[1]).read_text()
start=s.index('build_census() {\n'); end=s.index('\n}\n',start)+3
Path(sys.argv[2]).write_text(s[start:end])
PY
source "$scratch/build-function.sh"
build_census "$source" "$destination" "$logs"
sha256sum "$destination/census" "$destination/latent-census" "$destination/entry-census" > "$logs/binaries.sha256"
git -C "$source" status --porcelain --untracked-files=no > "$logs/compiler-status.txt"
test ! -s "$logs/compiler-status.txt"
(cd "$source/stage3/api" && npm ci --ignore-scripts --no-audit --no-fund) > "$logs/packages.log" 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-gate/stage3-meter.IuyyCt/upstream
bash "$source/stage3/apply.sh" "$scratch/adapted" > "$logs/apply.log" 2>&1
printf 'ready\n'
