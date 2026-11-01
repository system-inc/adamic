#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
cd /workspace/adamic
exec python3 stage3/census/changeb-road2-meter/65146f1e/run-censuses.py
