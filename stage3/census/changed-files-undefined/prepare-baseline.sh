#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/changed-files-proof/env.sh
cd /workspace/scratch/changed-files-typescript-603
git checkout --detach 050880ce59e30b356b686bd3144efe24f875ebc8
test "$(git rev-parse v6.0.3^{commit})" = 050880ce59e30b356b686bd3144efe24f875ebc8
node -p "require('./package.json').version"
cd /workspace/scratch/changed-files-undefined-1
stage3/apply.sh /workspace/scratch/changed-files-proof/main-tree > /workspace/scratch/changed-files-proof/main-apply.log 2>&1
