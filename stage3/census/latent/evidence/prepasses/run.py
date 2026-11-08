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
parser.add_argument('--peer-pid', action='append', type=int, default=[])
parser.add_argument('--rss-ceiling-kib', type=int)
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
requested_signal = None
peak_combined_rss = 0

def request_stop(number, frame):
    global requested_signal
    requested_signal = number

for number in (signal.SIGTERM, signal.SIGHUP, signal.SIGINT):
    signal.signal(number, request_stop)
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
            memory_stop = None
            if args.rss_ceiling_kib is not None:
                combined = 0
                for pid in [process.pid, *args.peer_pid]:
                    try:
                        status = Path('/proc',str(pid),'status').read_text()
                    except FileNotFoundError:
                        continue
                    for line in status.splitlines():
                        if line.startswith('VmRSS:'):
                            combined += int(line.split()[1])
                peak_combined_rss = max(peak_combined_rss,combined)
                if combined > args.rss_ceiling_kib:
                    memory_stop = 'combined VmRSS exceeded '+str(args.rss_ceiling_kib)+' KiB'
            if memory_stop is not None or requested_signal is not None or time.monotonic() - last_activity > 1800:
                stopped = memory_stop or (('runner received signal '+str(requested_signal)) if requested_signal is not None else '30 minutes without output or watched-file progress')
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
              user_seconds=usage.ru_utime, system_seconds=usage.ru_stime,
              peak_observed_combined_rss_kib=peak_combined_rss, rss_ceiling_kib=args.rss_ceiling_kib)
args.prefix.with_suffix('.status.json').write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result), flush=True)
raise SystemExit(code if code >= 0 else 128-code)
