"""Run one child with retained logs, own exit status and a 30-minute silence stop."""
import argparse
import json
import os
from pathlib import Path
import resource
import signal
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('prefix', type=Path)
parser.add_argument('--watch', action='append', default=[])
parser.add_argument('command', nargs=argparse.REMAINDER)
args = parser.parse_args()
args.prefix.parent.mkdir(parents=True, exist_ok=True)
log_path = args.prefix.with_suffix('.log')
command = args.command
if command and command[0] == '--':
    command = command[1:]
started = time.monotonic()
stopped = None
last_activity = started
previous = None
process = None
try:
    with log_path.open('w') as log:
        process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        while process.poll() is None:
            stamps = []
            for path in [log_path, *map(Path, args.watch)]:
                if path.exists():
                    stat = path.stat()
                    stamps.append((str(path), stat.st_mtime_ns, stat.st_size))
            if stamps != previous:
                last_activity, previous = time.monotonic(), stamps
            if time.monotonic() - last_activity > 1800:
                stopped = '30 minutes without output or watched-file progress'
                os.killpg(process.pid, signal.SIGTERM)
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                break
            time.sleep(1)
        code = process.wait()
finally:
    if process is not None and process.poll() is None:
        os.killpg(process.pid, signal.SIGKILL)
        process.wait()
usage = resource.getrusage(resource.RUSAGE_CHILDREN)
result = dict(command=command, child_exit=code, stopped=stopped,
              wall_seconds=time.monotonic()-started, peak_rss_kib=usage.ru_maxrss,
              user_seconds=usage.ru_utime, system_seconds=usage.ru_stime)
args.prefix.with_suffix('.status.json').write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result), flush=True)
raise SystemExit(code if code >= 0 else 128-code)
