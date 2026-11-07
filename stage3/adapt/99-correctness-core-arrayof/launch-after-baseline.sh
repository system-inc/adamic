#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
python3 /workspace/scratch/arrayof/prepare-measure.py > /workspace/scratch/arrayof/prepare-measure.log 2>&1
python3 /workspace/scratch/arrayof/count-pieces.py > /workspace/scratch/arrayof/count-pieces.log 2>&1
