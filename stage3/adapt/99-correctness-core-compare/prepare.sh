#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
while [ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse --verify HEAD 2>/dev/null || true)" != 050880ce59e30b356b686bd3144efe24f875ebc8 ]; do sleep 2; done
while [ ! -f /workspace/scratch/core-compare/api-install.log ]; do sleep 2; done
while ! node -e 'require("./stage3/api/node_modules/typescript")' >/dev/null 2>&1; do sleep 2; done
bash stage3/apply.sh --build-product /workspace/scratch/core-compare/before > /workspace/scratch/core-compare/apply-before.log 2>&1
printf '0\n' > /workspace/scratch/core-compare/apply-before.exit
cp -a /workspace/scratch/core-compare/before /workspace/scratch/core-compare/counted
node /workspace/scratch/core-compare/instrument.cjs /workspace/scratch/core-compare/counted > /workspace/scratch/core-compare/instrument.json
cd /workspace/scratch/core-compare/counted
npm ci --no-audit --no-fund > /workspace/scratch/core-compare/counted-install.log 2>&1
npm run build > /workspace/scratch/core-compare/counted-build.log 2>&1
npm exec hereby -- tests > /workspace/scratch/core-compare/counted-harness.log 2>&1
printf '0\n' > /workspace/scratch/core-compare/prepare.exit
