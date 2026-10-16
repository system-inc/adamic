#!/usr/bin/env python3
"""Interleave parent/current release binaries; run after the test gate is idle."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--before', required=True, help='binary prefix, e.g. /tmp/before')
parser.add_argument('--after', required=True, help='binary prefix, e.g. /tmp/after')
parser.add_argument('--threads', default='1,2,4')
parser.add_argument('--rounds', type=int, default=5)
parser.add_argument('--output', required=True)
args = parser.parse_args()
threads = sorted(set(int(x) for x in args.threads.split(',')))
cases = [('trees', 1)] + [(name, n) for name in ('files', 'share', 'map') for n in threads]
report = {'platform': platform.platform(), 'affinity': sorted(os.sched_getaffinity(0)) if hasattr(os, 'sched_getaffinity') else None,
          'cpu_max': Path('/sys/fs/cgroup/cpu.max').read_text().strip() if Path('/sys/fs/cgroup/cpu.max').exists() else None,
          'load_start': os.getloadavg(), 'rounds': args.rounds, 'samples': []}
outputs = {}
for round_index in range(args.rounds):
    order = cases[round_index:] + cases[:round_index]
    for name, n in order:
        versions = ('before', 'after') if round_index % 2 == 0 else ('after', 'before')
        for version in versions:
            binary = getattr(args, version) + '-' + name
            env = dict(os.environ, ADAMIC_THREADS=str(n))
            command = [binary] + (['million', 'time'] if name == 'map' else [])
            start = time.perf_counter()
            run = subprocess.run(command, env=env, capture_output=True, check=True)
            elapsed = time.perf_counter() - start
            sample = {'round': round_index + 1, 'program': name, 'threads': n, 'version': version, 'wall': elapsed}
            if name == 'share':
                sample.update(json.loads(run.stdout))
            else:
                if name not in outputs:
                    outputs[name] = run.stdout
                assert outputs[name] == run.stdout, (name, n, version, 'output mismatch')
                sample['stdout_sha256'] = hashlib.sha256(run.stdout).hexdigest()
            if name == 'map':
                sample['map'] = float(re.search(rb'map seconds ([0-9.]+)', run.stderr).group(1))
            report['samples'].append(sample)
            print(json.dumps(sample), flush=True)
report['load_end'] = os.getloadavg()
report['best'] = []
for name, n in cases:
    entry = {'program': name, 'threads': n}
    key = 'total' if name == 'share' else 'map' if name == 'map' else 'wall'
    for version in ('before', 'after'):
        samples = [s for s in report['samples'] if s['program'] == name and s['threads'] == n and s['version'] == version]
        values = [s[key] for s in samples]
        entry[version] = min(values)
        entry[version + '_spread_percent'] = (max(values) / min(values) - 1) * 100
        if name == 'share':
            entry[version + '_serial_share'] = min(s['share'] for s in samples)
            entry[version + '_parallel_section'] = min(s['map'] for s in samples)
    entry['change_percent'] = (entry['after'] / entry['before'] - 1) * 100
    report['best'].append(entry)
Path(args.output).write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps({'load_start': report['load_start'], 'load_end': report['load_end'], 'best': report['best']}, indent=2))
