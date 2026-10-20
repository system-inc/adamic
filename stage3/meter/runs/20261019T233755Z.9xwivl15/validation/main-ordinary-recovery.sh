#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/badrnrf-inputs/TypeScript
run=$(cd "$(dirname "$0")/.." && pwd)
scratch=/tmp/adamic-gate/stage3-meter.RXmuND
bash "$scratch/main-source/stage3/apply.sh" "$scratch/main-adapted" > "$run/main/apply.log" 2>&1
mkdir -p "$run/main/tsc"
"$scratch/main-binaries/census" "$scratch/main-adapted/src/tsc/tsc.ts" "$run/main/tsc/census.jsonl" > "$run/main/tsc/census.log" 2>&1
"$scratch/main-binaries/census" "$scratch/main-adapted/src/compiler" "$run/main/census.jsonl" > "$run/main/census.log" 2>&1
