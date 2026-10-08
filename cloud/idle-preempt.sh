#!/usr/bin/env bash
set -euo pipefail
# BEGIN idle preemption (keep identical to cloud/idle-preempt.sh)
python3 - <<'IDLE_PREEMPT'
import datetime, os, signal, time

started = time.monotonic()
deadline = started + 10
sessions, found, survivors = set(), set(), {}

def processes():
    result = {}
    for name in os.listdir('/proc'):
        if not name.isdigit():
            continue
        try:
            path = '/proc/' + name
            with open(path + '/stat') as handle:
                fields = handle.read().rsplit(')', 1)[1].split()
            # Fields after comm start at stat field 3: session is 6, starttime 22.
            if fields[0] not in ('Z', 'X'):
                result[int(name)] = (int(fields[3]), fields[19], os.stat(path).st_uid)
        except (OSError, ProcessLookupError):
            pass
    return result

def marked():
    result = {}
    for pid, info in processes().items():
        if info[2] != os.getuid():
            continue
        try:
            with open('/proc/%d/environ' % pid, 'rb') as handle:
                is_marked = b'ADAMIC_IDLE_JOB=1' in handle.read().split(b'\0')
            if is_marked:
                result[pid] = info
                sessions.add(info[0])
                found.add((pid, info[1]))
        except (OSError, ProcessLookupError):
            pass
    return result

def in_sessions():
    return {pid: info for pid, info in processes().items() if info[0] in sessions}

def kill(pids):
    current = processes()
    for pid, info in pids.items():
        # Do not signal an unrelated process if a PID has been reused.
        if current.get(pid) != info:
            continue
        try:
            os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass

def disable(pids):
    new = {pid: info for pid, info in pids.items() if (pid, info[1]) not in survivors}
    if not new:
        return
    directory = os.path.expanduser('~/idle')
    os.makedirs(directory, exist_ok=True)
    with open(directory + '/disabled', 'a') as report:
        report.write(datetime.datetime.now(datetime.timezone.utc).isoformat() +
                     ' idle session survivors; human review required\n')
        for pid, info in sorted(new.items()):
            try:
                with open('/proc/%d/cmdline' % pid, 'rb') as handle:
                    command = handle.read().replace(b'\0', b' ').decode(errors='replace')
            except OSError:
                command = '<exited before command line could be read>'
            report.write('pid=%d session=%d command=%r\n' % (pid, info[0], command))
            survivors[(pid, info[1])] = info
        report.flush()
        os.fsync(report.fileno())

try:
    # Kill marked parents before they can spawn more children; repeat for races.
    while True:
        pids = marked()
        if not pids:
            break
        kill(pids)
        if time.monotonic() >= deadline:
            break
        time.sleep(.05)
    # Check every live process in each recorded session, even without the marker.
    # A survivor disables further idle launches before we attempt to kill it.
    while True:
        pids = in_sessions()
        disable(pids)
        kill(pids)
        remaining = marked()
        remaining.update(in_sessions())
        if not remaining:
            break
        disable(remaining)
        if time.monotonic() >= deadline:
            raise SystemExit('idle preemption timed out; gate aborted')
        time.sleep(.05)
finally:
    elapsed = int((time.monotonic() - started) * 1000)
    print('idle preemption: %d marked, %d survivors in their sessions, %d ms' %
          (len(found), len(survivors), elapsed), flush=True)
IDLE_PREEMPT
# END idle preemption
