#!/usr/bin/env bash
set -euo pipefail
# BEGIN idle preemption (keep identical to cloud/idle-preempt.sh)
python3 - <<'IDLE_PREEMPT'
import os, signal, time

def marked():
    result = []
    for name in os.listdir('/proc'):
        if not name.isdigit():
            continue
        try:
            path = '/proc/' + name
            if os.stat(path).st_uid != os.getuid():
                continue
            if b'ADAMIC_IDLE_JOB=1' in open(path + '/environ', 'rb').read().split(b'\0'):
                if open(path + '/stat').read().rsplit(')', 1)[1].split()[0] != 'Z':
                    result.append(int(name))
        except (OSError, ProcessLookupError):
            pass
    return result

# SIGKILL prevents a dying parent from spawning new children after the scan.
# Repeat for children born between scanning /proc and killing their parent.
deadline = time.monotonic() + 10
while True:
    pids = marked()
    if not pids:
        break
    for pid in pids:
        try:
            os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    if time.monotonic() >= deadline:
        raise SystemExit('idle preemption timed out; gate aborted')
    time.sleep(.05)
IDLE_PREEMPT
# END idle preemption
