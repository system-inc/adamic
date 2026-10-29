#!/usr/bin/env python3
"""Record the meter host and its observed load before and after measurement."""
import json
from pathlib import Path
import os
import shutil
import socket
import subprocess
import sys


def loadavg():
    path = Path('/proc/loadavg')
    return path.read_text().strip() if path.exists() else subprocess.check_output(['uptime'], text=True).strip()


def cpu_quota():
    # cgroup v2, then v1. Keep quota and period together, in microseconds.
    path = Path('/sys/fs/cgroup/cpu.max')
    if path.exists():
        quota, period = path.read_text().split()
        return None if quota == 'max' else {'quota_us': int(quota), 'period_us': int(period)}
    for directory in ('/sys/fs/cgroup/cpu', '/sys/fs/cgroup/cpu,cpuacct'):
        path = Path(directory)
        if (path / 'cpu.cfs_quota_us').exists():
            quota = int((path / 'cpu.cfs_quota_us').read_text())
            return None if quota < 0 else {'quota_us': quota, 'period_us': int((path / 'cpu.cfs_period_us').read_text())}
    return None


def record(path, phase):
    if phase == 'start':
        memory = Path('/proc/meminfo')
        total = next(line.split(':', 1)[1].strip() for line in memory.read_text().splitlines()
                     if line.startswith('MemTotal:')) if memory.exists() else subprocess.check_output(['sysctl', '-n', 'hw.memsize'], text=True).strip() + ' bytes'
        box = {'hostname': socket.gethostname(),
               'nproc': int(subprocess.check_output(['nproc'], text=True)) if shutil.which('nproc') else os.cpu_count(),
               'cpu_quota': cpu_quota(), 'total_memory': total, 'loadavg_start': loadavg()}
    elif phase == 'end':
        box = json.loads(path.read_text())
        box['loadavg_end'] = loadavg()
    else:
        raise ValueError('expected start or end')
    path.write_text(json.dumps(box, indent=2) + '\n')


if __name__ == '__main__':
    record(Path(sys.argv[1]), sys.argv[2])
