#!/usr/bin/env bash
# Delivered over ssh by idle-jobs.py. All descendants inherit the exact marker.
set -euo pipefail
mode=$1
shift
locks() {
  local path fd=30
  for path in "$HOME/fast-gate/lock" "$HOME/fast-gate/lock-2" "$HOME/fast-gate/lock-3" "$HOME/full-gate/lock"; do
    # In read-only probes absent locks mean no gate has created them yet.
    if [ "$mode" = probe ]; then
      [ -e "$path" ] || continue
      eval "exec $fd<\"\$path\""
    else
      mkdir -p "$(dirname "$path")"
      eval "exec $fd>\"\$path\""
    fi
    flock -n "$fd" || return 1
    fd=$((fd + 1))
  done
}
if [ "$mode" = probe ]; then
  if ! locks; then echo gated; exit; fi
  if [ -e ~/idle/lock ]; then
    exec 8< ~/idle/lock
    flock -n 8 || { echo 'running a job'; exit; }
  fi
  echo idle
  exit
fi
sha=$1 seed=$2 count=$3 box=$4
mkdir -p ~/idle
exec 8> ~/idle/lock
flock -n 8 || { echo running; exit; }
locks || { echo gated; exit; }
# Hold admission locks until the marked child acknowledges that it is alive.
# The job holds only the idle lock; a gate never waits for its batch.
cat > ~/idle/job.sh
ready="$HOME/idle/ready-$seed"
rm -f "$ready"
nohup env ADAMIC_IDLE_JOB=1 nice -n 19 bash ~/idle/job.sh "$sha" "$seed" "$count" "$box" "$ready" 30>&- 31>&- 32>&- 33>&- </dev/null >~/idle/driver.log 2>&1 &
pid=$!
for ((i=0; i<100; i++)); do
  if [ -f "$ready" ]; then echo started; exit; fi
  kill -0 "$pid" 2>/dev/null || { echo failed; exit 1; }
  sleep .01
done
kill "$pid" 2>/dev/null || true
echo 'idle startup timed out' >&2
exit 1
