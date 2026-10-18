#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4
export ADAMIC_GATE_UNCACHED=1
export GOPROXY='https://proxy.golang.org|direct'
evidence=review/compiler/loom-slice-2-2
timeout 600 python3 "$evidence/run-containment-mutants.py" > "$evidence/containment-runner.log" 2>&1
timeout 360 python3 "$evidence/run-callback-mutants.py" > "$evidence/callback-runner.log" 2>&1
timeout 150 python3 "$evidence/run-weak-isolated.py" > "$evidence/weak-isolated-runner.log" 2>&1
timeout 300 python3 "$evidence/run-split-mutants.py" > "$evidence/split-runner.log" 2>&1
