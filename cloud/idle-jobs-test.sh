#!/usr/bin/env bash
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
tmp=$(mktemp -d)
marked='' unmarked=''
cleanup() {
  [ -z "$marked" ] || kill "$marked" 2>/dev/null || true
  [ -z "$unmarked" ] || kill "$unmarked" 2>/dev/null || true
  rm -rf "$tmp"
}
trap cleanup EXIT
bash -n "$here"/idle*.sh "$here/fast-gate.sh" "$here/full-gate-main.sh"
# Exercise the exact inline preemption block used by fast-gate.sh.
sed -n '/^# BEGIN idle preemption/,/^# END idle preemption/p' "$here/fast-gate.sh" > "$tmp/preempt.sh"
sed -n '/^# BEGIN idle preemption/,/^# END idle preemption/p' "$here/idle-preempt.sh" > "$tmp/expected.sh"
cmp "$tmp/preempt.sh" "$tmp/expected.sh"
env ADAMIC_IDLE_JOB=1 sleep 120 &
marked=$!
sleep 120 &
unmarked=$!
sleep .1
exec 9> "$tmp/fake-gate-lock"
flock 9
bash "$tmp/preempt.sh"
if kill -0 "$marked" 2>/dev/null; then echo 'marked process survived' >&2; exit 1; fi
kill -0 "$unmarked"
wait "$marked" 2>/dev/null || true
marked=''
echo 'PASS: fast gate took fake lock; marked sleep died; unmarked sleep survived'
mkdir -p "$tmp/bin" "$tmp/state"
touch "$tmp/state/enabled"
cat > "$tmp/bin/ssh" <<'SSH'
#!/usr/bin/env bash
set -euo pipefail
box=$1
# Never execute the input: verify dry-run sends only the read-only probe.
[[ $2 == 'bash -s -- probe' ]]
cat > "$IDLE_TEST_TMP/$box.probe"
grep -q 'if \[ "$mode" = probe \]' "$IDLE_TEST_TMP/$box.probe"
case "$box" in cloud) echo gated;; *) echo idle;; esac
SSH
cat > "$tmp/bin/git" <<'GIT'
#!/usr/bin/env bash
printf '%s\trefs/heads/main\n' 0123456789012345678901234567890123456789
GIT
chmod +x "$tmp/bin/ssh" "$tmp/bin/git"
PATH="$tmp/bin:$PATH" IDLE_TEST_TMP="$tmp" ADAMIC_IDLE_STATE="$tmp/state" bash "$here/idle-jobs.sh" --dry-run > "$tmp/dry.log"
cat "$tmp/dry.log"
grep -q '^cloud: gated.*would start no job$' "$tmp/dry.log"
grep -q '^home: idle.*would start seed=1 count=2000$' "$tmp/dry.log"
grep -q '^chonchon: idle.*would start seed=6001 count=2000$' "$tmp/dry.log"
[ ! -e "$tmp/state/state.json" ]
[ ! -e "$tmp/state/coordinator.lock" ]
rm "$tmp/state/enabled"
PATH="$tmp/bin:$PATH" IDLE_TEST_TMP="$tmp" ADAMIC_IDLE_STATE="$tmp/state" bash "$here/idle-jobs.sh" --dry-run > "$tmp/disabled.log"
! grep -q 'would start seed=' "$tmp/disabled.log"
echo 'PASS: fake SSH dry-run, seed allocation, and disabled switch; no state writes'
# Real local flock probe: absent locks, then a held third fast slot.
mkdir -p "$tmp/box/fast-gate" "$tmp/box/full-gate"
[ "$(HOME="$tmp/box" bash "$here/idle-job-box.sh" probe)" = idle ]
exec 10> "$tmp/box/fast-gate/lock-3"
flock 10
[ "$(HOME="$tmp/box" bash "$here/idle-job-box.sh" probe)" = gated ]
[ "$(HOME="$tmp/box" bash "$here/idle-job-box.sh" start 0123456789012345678901234567890123456789 1 2000 test)" = gated ]
[ ! -e "$tmp/box/idle/job.sh" ]
echo 'PASS: third-slot lock blocks admission before any worker starts'
# Launch the real admission shell with a tiny marked worker, then preempt it.
flock -u 10
python3 - "$here" "$tmp/launch.sh" <<'PY'
import pathlib, sys
source = (pathlib.Path(sys.argv[1]) / 'idle-job-box.sh').read_text()
worker = 'touch "$5"\nsleep 120\n'
source = source.replace('cat > ~/idle/job.sh', "cat > ~/idle/job.sh <<'WORKER'\n" + worker + 'WORKER')
pathlib.Path(sys.argv[2]).write_text(source)
PY
[ "$(HOME="$tmp/box" bash "$tmp/launch.sh" start 0123456789012345678901234567890123456789 1 2000 test)" = started ]
[ "$(HOME="$tmp/box" bash "$here/idle-job-box.sh" probe)" = 'running a job' ]
exec 11> "$tmp/box/fast-gate/lock"
flock -n 11
bash "$tmp/preempt.sh"
[ "$(HOME="$tmp/box" bash "$here/idle-job-box.sh" probe)" = gated ]
flock -u 11
[ "$(HOME="$tmp/box" bash "$here/idle-job-box.sh" probe)" = idle ]
kill -0 "$unmarked"
echo 'PASS: admitted worker releases gate locks, retains idle lock, and preemption removes its children'
