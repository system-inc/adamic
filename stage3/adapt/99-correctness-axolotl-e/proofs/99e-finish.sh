#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
while [ ! -f /workspace/scratch/99e-oracle-after/report.json ]; do sleep 2; done
PYTHONDONTWRITEBYTECODE=1 python3 stage3/adapt/99-correctness-axolotl-e/proofs/verify-oracles.py /workspace/scratch /workspace/scratch/99e-verification > /workspace/scratch/99e-verification.log 2>&1
node stage3/adapt/99-correctness-axolotl-e/proofs/witness.cjs /workspace/scratch/99e-before/built/local/typescript.js /workspace/scratch/99e-after/built/local/typescript.js > /workspace/scratch/99e-witness.log 2>&1
python3 stage3/adapt/99-correctness-axolotl-e/proofs/native.py /workspace/scratch/99e-tsgo > /workspace/scratch/99e-native.json 2>&1
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-e/proofs/check.cjs /workspace/scratch/99e-before /workspace/scratch/99e-after > /workspace/scratch/99e-contract-final.log 2>&1
python3 stage3/adapt/99-correctness-axolotl-e/proofs/scope.py /workspace/scratch/99e-before /workspace/scratch/99e-after /workspace/scratch/99e-scope-final > /workspace/scratch/99e-scope-final.log 2>&1
