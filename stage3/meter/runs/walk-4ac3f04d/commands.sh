#!/usr/bin/env bash
# Command record, not an entry point. Run only the desired block.
# Repository: /workspace/adamic. All measurements ran in foreground tool sessions.
# The first setup was started on the supplied work checkout; switching branches
# during its build caused an incoherent source file list. Both setup logs are kept.
cd /workspace/adamic
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1
setup_exit=$?
echo "exit=$setup_exit"
if [ "$setup_exit" -eq 0 ]; then git fetch origin main; fi

# Independent fetch/read of the requested earlier counting script.
git fetch origin typescript/stage3-meter-daily-20261026-tzd3gjg:refs/remotes/origin/typescript/stage3-meter-daily-20261026-tzd3gjg
git show origin/typescript/stage3-meter-daily-20261026-tzd3gjg:stage3/meter/runs/20261025T234850Z.FeRsFs/count-predicates.py
git fetch --no-recurse-submodules origin main
git log --oneline 76ea8b17ea0ed467ecb86f355c06f738fd262983..4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288
git show -s --format='%H %P %s' 4ac3f04d
git diff --stat 4ac3f04d -- stage3/meter/twice-daily.sh stage3/census/latent
git diff --name-only 76ea8b17ea0ed467ecb86f355c06f738fd262983 4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288 -- stage3 cohere

# This switch overlapped the first setup build; the setup rerun below was stable.
git switch -c typescript/stage3-meter-walk-4ac3f04d 4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288
mkdir -p stage3/meter/runs/walk-4ac3f04d
source /workspace/adamic-tools/env.sh
{ date -u +%FT%TZ; nproc; cat /sys/fs/cgroup/cpu.max; free -h; free -b; node --version; go version; clang --version; uname -a; } > stage3/meter/runs/walk-4ac3f04d/box.txt

# Repeat setup after the switch; exit=0. No branch switches occurred afterward.
cp /tmp/setup.log stage3/meter/runs/walk-4ac3f04d/setup-initial.log
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1
setup_exit=$?
echo "exit=$setup_exit"
if [ "$setup_exit" -eq 0 ]; then git fetch origin main; fi

# The exact measurement loop, with explicit failure stop added to the requested loop.
cp /tmp/setup.log stage3/meter/runs/walk-4ac3f04d/setup.log
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
for sha in 76ea8b17ea0ed467ecb86f355c06f738fd262983 4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288; do
  started=$(date -u +%FT%TZ)
  started_seconds=$(date +%s)
  STAGE3_METER_MAIN_REF=$sha STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error \
    timeout --kill-after=60s 7200 bash stage3/meter/twice-daily.sh > /tmp/meter-$sha.log 2>&1
  meter_exit=$?
  elapsed_seconds=$(($(date +%s)-started_seconds))
  echo "$sha exit=$meter_exit"
  printf '%s %s %s %s\n' "$sha" "$started" "$elapsed_seconds" "$meter_exit" >> stage3/meter/runs/walk-4ac3f04d/run-status.txt
  cp /tmp/meter-$sha.log stage3/meter/runs/walk-4ac3f04d/meter-$sha.log
  if [ "$meter_exit" -ne 0 ]; then break; fi
done

# While the before run was in its measurement phase, after all its builds and
# adaptations had finished, these generated inputs were removed to free tmpfs.
# Retained: binaries, overlays, adapted trees, and every run artifact/log.
git worktree remove --force --force /tmp/adamic-gate/stage3-meter.nfa1mc/main-source
git worktree remove --force --force /tmp/adamic-gate/stage3-meter.nfa1mc/area-source
rm -rf /tmp/adamic-gate/stage3-meter.nfa1mc/upstream
df -h /tmp

# Equivalent cleanup after both after-run builds/adaptations finished.
git worktree remove --force --force /tmp/adamic-gate/stage3-meter.4OGO7t/main-source
git worktree remove --force --force /tmp/adamic-gate/stage3-meter.4OGO7t/area-source
rm -rf /tmp/adamic-gate/stage3-meter.4OGO7t/upstream

# Initial timing prerequisite attempt: exit 127; latent was not launched.
# time-checker.sh originally used /usr/bin/time; that exact attempted invocation
# is in timing/prerequisite-failure/commands.log. Its stderr is retained there.
bash stage3/meter/runs/walk-4ac3f04d/time-checker.sh > /tmp/walk-checker-timing.log 2>&1
timing_exit=$?
echo "timing exit=$timing_exit"
cp /tmp/walk-checker-timing.log stage3/meter/runs/walk-4ac3f04d/timing.log

# System install failed (non-root Unix permissions), apt download had no indexes.
apt-get update > /tmp/walk-time-install.log 2>&1 && apt-get install -y time >> /tmp/walk-time-install.log 2>&1
install_exit=$?
echo "GNU time install exit=$install_exit"
mkdir -p /workspace/adamic-tools/gnu-time
cd /workspace/adamic-tools/gnu-time
apt-get download time > /tmp/walk-time-download.log 2>&1
echo "time download exit=$?"
curl -fsSL https://deb.debian.org/debian/pool/main/t/time/time_1.9-0.2_amd64.deb -o time_1.9-0.2_amd64.deb
dpkg-deb -x time_1.9-0.2_amd64.deb extracted
extracted/usr/bin/time --version
sha256sum time_1.9-0.2_amd64.deb extracted/usr/bin/time
cd /workspace/adamic

# The successful timing script uses the extracted package's usr/bin/time -v.
# The six fully expanded commands are in timing/commands.log.
bash stage3/meter/runs/walk-4ac3f04d/time-checker.sh > /tmp/walk-checker-timing.log 2>&1
timing_exit=$?
echo "timing exit=$timing_exit"
cp /tmp/walk-checker-timing.log stage3/meter/runs/walk-4ac3f04d/timing.log

# Regenerate every comparison table from committed raw evidence.
python3 stage3/meter/runs/walk-4ac3f04d/compare.py > stage3/meter/runs/walk-4ac3f04d/changed-reasons.md

# Preserve the earlier counting logic reused by compare.py.
git show origin/typescript/stage3-meter-daily-20261026-tzd3gjg:stage3/meter/runs/20261025T234850Z.FeRsFs/count-predicates.py > stage3/meter/runs/walk-4ac3f04d/count-predicates-original.py.txt

# Validation and publication (only the requested branch; no force push).
python3 stage3/meter/runs/walk-4ac3f04d/compare.py > stage3/meter/runs/walk-4ac3f04d/changed-reasons.md
git diff --check
git add -f stage3/meter/runs/20261027T014837Z.mugwl6 stage3/meter/runs/20261027T025412Z.HCmJYO stage3/meter/runs/walk-4ac3f04d
git diff --cached --check
git diff --cached --stat
git commit -m "Record before/after Stage 3 meter for 4ac3f04d false-direction walk"
git -c push.followTags=false push -u origin typescript/stage3-meter-walk-4ac3f04d
git rev-parse HEAD
git ls-remote --heads origin refs/heads/typescript/stage3-meter-walk-4ac3f04d
git status --short --branch
