#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
while [[ ! -f /workspace/scratch/20pb346-2/after-compiler-cap/run.json ]]; do sleep 1; done
python3 stage3/meter/evidence/unit-visibility-20pb346/completed-pass-proof.py /workspace/scratch/20pb346-2 /workspace/scratch/20pb346-2/completed-fixture > /workspace/scratch/20pb346-2-completed-fixture.log 2>&1
CENSUS_BINARY=/workspace/scratch/20pb346-2/census ENTRY_CENSUS_BINARY=/workspace/scratch/20pb346-2/before-entry LATENT_CENSUS_BINARY=/workspace/scratch/20pb346-2/before-compiler python3 -E -S -B -m unittest discover -s /workspace/scratch/20pb346-2/main/stage3/meter -p '*_test.py' > /workspace/scratch/20pb346-2-main-python-real-tests.log 2>&1
CENSUS_BINARY=/workspace/scratch/20pb346-2/census ENTRY_CENSUS_BINARY=/workspace/scratch/20pb346-2/after-entry LATENT_CENSUS_BINARY=/workspace/scratch/20pb346-2/after-compiler python3 -E -S -B -m unittest discover -s stage3/meter -p '*_test.py' > /workspace/scratch/20pb346-2-after-python-real-tests.log 2>&1
python3 /workspace/scratch/20pb346-2/run.py before compiler before-compiler-complete 10800 > /workspace/scratch/20pb346-2-before-compiler-complete.log 2>&1
python3 /workspace/scratch/20pb346-2/run.py after compiler after-compiler-complete 10800 > /workspace/scratch/20pb346-2-after-compiler-complete.log 2>&1
