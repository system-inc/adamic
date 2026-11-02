#!/usr/bin/env bash
# Material commands used during the 99c run. Full output was redirected to files.
# Baseline apply began before 99-correctness-axolotl-c existed.
# For replay, run baseline apply from a checkout of main at the recorded base.
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
source /workspace/adamic-tools/env.sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /workspace/scratch/upstream > /workspace/scratch/upstream.log 2>&1
(cd stage3/api && npm ci --ignore-scripts --no-audit --no-fund > /workspace/scratch/packages.log 2>&1)
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/before > /workspace/scratch/apply-before.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/after > /workspace/scratch/apply-after.log 2>&1
(cd /workspace/scratch/before && npm ci --no-audit --no-fund > /workspace/scratch/install-before.log 2>&1 && npm run build > /workspace/scratch/build-before.log 2>&1)
NODE_PATH=/workspace/adamic/stage3/api/node_modules node stage3/lane/generate-api.cjs /workspace/scratch/upstream > /workspace/scratch/api-expected.json 2> /workspace/scratch/api-expected.stderr
bash stage3/oracle/run.sh /workspace/scratch/after /workspace/scratch/oracle > /workspace/scratch/oracle.log 2>&1
# Assert generated and local API equal generator's expected bytes, assert only API
# changed in the raw oracle, then copy those proven bytes to external reference.
bash stage3/oracle/run.sh /workspace/scratch/after /workspace/scratch/oracle-accepted > /workspace/scratch/oracle-accepted.log 2>&1
NODE_PATH=/workspace/adamic/stage3/api/node_modules node /workspace/scratch/verify.cjs > /workspace/scratch/verify.json 2> /workspace/scratch/verify.stderr
node /workspace/scratch/own-keys.cjs > /workspace/scratch/own-keys.json 2> /workspace/scratch/own-keys.stderr
go build -o /workspace/scratch/adamic ./cmd/adamic > /workspace/scratch/build-adamic.log 2>&1
python3 /workspace/scratch/check-types.py > /workspace/scratch/check-types.log 2>&1
python3 /workspace/scratch/mutants.py > /workspace/scratch/mutants.log 2>&1

# The ten-member probe exposed TSConfig.exclude; the exclude-only probe exposed
# TSConfig.include. The final adapter contains both internal sibling unions.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/after-final > /workspace/scratch/apply-after-final.log 2>&1
(cd /workspace/scratch/after-final && npm ci --no-audit --no-fund > /workspace/scratch/install-after-final.log 2>&1)
python3 /workspace/scratch/check-types.py > /workspace/scratch/check-types.log 2>&1
python3 /workspace/scratch/mutants.py > /workspace/scratch/mutants.log 2>&1
(cd /workspace/scratch/after-final && npm run build > /workspace/scratch/build-after-final.log 2>&1)
NODE_PATH=/workspace/adamic/stage3/api/node_modules node /workspace/scratch/verify-final.cjs > /workspace/scratch/verify-complete.json 2> /workspace/scratch/verify-complete.stderr
node /workspace/scratch/own-keys.cjs > /workspace/scratch/own-keys-final.json 2> /workspace/scratch/own-keys-final.stderr
# After the exact API guard in api-final-proof.json, accept only that reference.
bash stage3/oracle/run.sh /workspace/scratch/after-final /workspace/scratch/oracle-final > /workspace/scratch/oracle-final.log 2>&1
