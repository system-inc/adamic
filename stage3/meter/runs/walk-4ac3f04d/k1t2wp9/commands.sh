#!/usr/bin/env bash
# Command record, not an entry point. Blocks were executed separately in foreground.
cd /workspace/adamic
git fetch origin typescript/stage3-meter-walk-4ac3f04d
git checkout -b typescript/stage3-meter-walk-4ac3f04d FETCH_HEAD
# Read COMPARISON.md, commands.sh, compare.py, time-checker.sh and twice-daily.sh first.
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1
setup_exit=$?
echo "exit=$setup_exit"
if [ "$setup_exit" -eq 0 ]; then git fetch origin main; fi
source /workspace/adamic-tools/env.sh
out=stage3/meter/runs/walk-4ac3f04d/k1t2wp9
mkdir -p "$out"
cp /tmp/setup.log "$out/setup.log"
{ date -u +%FT%TZ; nproc; cat /sys/fs/cgroup/cpu.max; free -b; node --version; go version; clang --version; } > "$out/box.txt"
started_seconds=$(date +%s)
STAGE3_METER_MAIN_REF=8502a46e9289d6abb673f1359aa12d66be659b55 STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error timeout --kill-after=60s 7200 bash stage3/meter/twice-daily.sh > /tmp/meter-8502.log 2>&1
meter_exit=$?
cp /tmp/meter-8502.log "$out/meter-8502.log"
printf 'exit=%s elapsed=%s\n' "$meter_exit" "$(($(date +%s)-started_seconds))" > "$out/run-status.txt"
run=$(sed -n 's/^run: //p' /tmp/meter-8502.log | head -1)
printf '%s\n' "$run" > "$out/run-path.txt"
if [ -n "$run" ]; then git add -f "$run"; fi
git add -f "$out"
git commit -m 'Preserve single 8502a46e Stage 3 meter run and logs' && git -c push.followTags=false push origin HEAD:refs/heads/typescript/stage3-meter-walk-4ac3f04d
echo "meter exit=$meter_exit"
# While the meter was active, after both builds and adaptations completed:
# git worktree remove --force --force /tmp/adamic-gate/stage3-meter.IuyyCt/main-source
# git worktree remove --force --force /tmp/adamic-gate/stage3-meter.IuyyCt/area-source
# cat /sys/fs/cgroup/memory.max >> "$out/box.txt"
# The report-only comparison and timing scripts were prepared during the run.
# No timing samples or before rebuild overlapped the meter.
python3 stage3/meter/runs/walk-4ac3f04d/compare.py --k1t2wp9
started=$(date +%s)
timeout --kill-after=60s 2700 bash "$out/build-before.sh" > "$out/before-build-driver.log" 2>&1
status=$?
printf 'exit=%s elapsed=%s\n' "$status" "$(($(date +%s)-started))" > "$out/before-build-status.txt"
bash "$out/time-checker.sh" > "$out/timing-driver.log" 2>&1
status=$?
printf 'exit=%s\n' "$status" > "$out/timing-status.txt"
python3 stage3/meter/runs/walk-4ac3f04d/compare.py --k1t2wp9
# Detailed expanded timing commands are in timing/commands.log.
# Final validation and publication:
python3 stage3/meter/runs/walk-4ac3f04d/compare.py
git diff --check
git add -f stage3/meter/runs/walk-4ac3f04d/compare.py stage3/meter/runs/walk-4ac3f04d/K1T2WP9.md "$out"
git diff --cached --check
git commit -m 'Compare K1T2WP9 meter with 4ac3f04d and record six timing samples'
git -c push.followTags=false push origin HEAD:refs/heads/typescript/stage3-meter-walk-4ac3f04d
