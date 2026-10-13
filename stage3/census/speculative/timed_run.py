"""Run one measurement with time, address-space and resident-memory bounds.

Usage: timed_run.py SECONDS RSS_MIB METRICS_JSON LOG COMMAND [ARG ...]
GNU timeout supplies the wall deadline. The supervisor samples the process group
RSS every 100ms and kills the group on the first sample above the limit. RLIMIT_AS
also caps each process at twice the RSS allowance. GOMEMLIMIT is only advisory.
"""
import json
import os
from pathlib import Path
import resource
import signal
import subprocess
import sys
import time


def resident_bytes(group):
    total = 0
    for path in Path('/proc').glob('[0-9]*/stat'):
        try:
            stat = path.read_text().rsplit(')', 1)[1].split()
            if int(stat[2]) == group:
                total += int(stat[21]) * os.sysconf('SC_PAGE_SIZE')
        except (OSError, ValueError, IndexError):
            pass
    return total


def run(seconds, rss_mib, metrics, log, command):
    limit = int(rss_mib) * 1024 * 1024
    def bounded():
        os.setsid()
        resource.setrlimit(resource.RLIMIT_AS, (2 * limit, 2 * limit))
    started = time.monotonic()
    peak = 0
    killed = False
    with Path(log).open('wb') as stream:
        process = subprocess.Popen(['timeout', '-k', '2s', str(seconds) + 's', *command],
                                   stdout=stream, stderr=subprocess.STDOUT, preexec_fn=bounded)
        while process.poll() is None:
            peak = max(peak, resident_bytes(process.pid))
            if peak > limit:
                killed = True
                os.killpg(process.pid, signal.SIGKILL)
                break
            time.sleep(0.1)
        code = process.wait()
    usage = resource.getrusage(resource.RUSAGE_CHILDREN)
    result = dict(command=command, wall_seconds=time.monotonic() - started,
                  peak_rss_kib=max(peak // 1024, usage.ru_maxrss), exit=code,
                  time_limit_seconds=float(seconds), rss_limit_mib=int(rss_mib),
                  address_space_limit_mib=2 * int(rss_mib), rss_limit_killed=killed)
    Path(metrics).write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result), flush=True)
    return code


if __name__ == '__main__':
    sys.exit(run(float(sys.argv[1]), int(sys.argv[2]), sys.argv[3], sys.argv[4], sys.argv[5:]))
