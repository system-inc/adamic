"""Recover an orphaned census's own Linux wait status after a transport reconnect."""
import json
import os
from pathlib import Path
import signal
import sys
import time

pid = int(sys.argv[1])
prefix = Path(sys.argv[2])
watch = Path(sys.argv[3])
proc = Path('/proc') / str(pid)
command = (proc/'cmdline').read_bytes().split(b'\0')
assert command[0], 'empty process command'
fields = (proc/'stat').read_text().rsplit(')',1)[1].split()
start_ticks = int(fields[19])
hz = os.sysconf('SC_CLK_TCK')
peak = 0
previous = None
last = time.monotonic() - max(0,time.time()-watch.stat().st_mtime)
stopped = None
term_at = None
while True:
    if not proc.exists():
        code = None
        stopped = stopped or 'process disappeared before its kernel wait status could be read'
        break
    fields = (proc/'stat').read_text().rsplit(')',1)[1].split()
    assert int(fields[19]) == start_ticks, 'PID reused'
    if fields[0] == 'Z':
        code = os.waitstatus_to_exitcode(int(fields[49]))
        break
    for line in (proc/'status').read_text().splitlines():
        if line.startswith('VmHWM:'):
            peak = max(peak,int(line.split()[1]))
    stat = watch.stat()
    stamp = (stat.st_mtime_ns,stat.st_size)
    if previous is not None and stamp != previous:
        last = time.monotonic()
    previous = stamp
    if stopped is None and time.monotonic()-last > 1800:
        stopped = '30 minutes without watched-unit progress'
        os.killpg(pid,signal.SIGTERM)
        term_at = time.monotonic()
    if term_at is not None and time.monotonic()-term_at > 10:
        os.killpg(pid,signal.SIGKILL)
        term_at = None
    time.sleep(1)
uptime = float(Path('/proc/uptime').read_text().split()[0])
result = dict(command=[part.decode() for part in command if part], child_exit=code,
              stopped=stopped, wall_seconds=uptime-start_ticks/hz,
              peak_rss_kib=peak, user_seconds=int(fields[11])/hz,
              system_seconds=int(fields[12])/hz,
              exit_status_source='Linux /proc/PID/stat exit_code after zombie state',
              resource_source='process lifetime CPU ticks and sampled lifetime VmHWM; final sampling interval may miss a peak',
              recovery='Original outer runner ended during environment reconnect; measurement process continued unchanged.',
              pid=pid)
prefix.with_suffix('.status.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
raise SystemExit(code if code is not None and code >= 0 else 1)
