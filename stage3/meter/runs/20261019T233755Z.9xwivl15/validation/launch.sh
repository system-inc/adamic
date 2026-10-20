#!/usr/bin/env bash
set -uo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
run=$(cd "$(dirname "$0")/.." && pwd)
started=$(date +%s)
printf '%s\n' "$started" > "$run/validation/meter-start.epoch"
date -u +%FT%TZ > "$run/validation/meter-start.utc"
STAGE3_METER_MAIN_REF=295df3bc6aa3d47933ccf675da1c33e745856313 \
STAGE3_METER_AREA_REF=295df3bc6aa3d47933ccf675da1c33e745856313 \
STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error \
STAGE3_METER_RUNS=$PWD/stage3/meter/runs \
timeout --kill-after=60s 14400 bash "$run/validation/pinned-twice-daily.sh" > "$run/meter-start.log" 2>&1
status=$?
finished=$(date +%s)
printf '%s\n' "$status" > "$run/validation/meter.exit"
printf '%s\n' "$((finished-started))" > "$run/validation/meter-elapsed.seconds"
date -u +%FT%TZ > "$run/validation/meter-end.utc"
exit "$status"
