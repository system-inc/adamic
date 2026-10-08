#!/usr/bin/env python3
"""Five checksum-checked rounds; builds, validation and profiling are untimed."""
import argparse
import json
import os
import platform
import resource
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--before', required=True)
parser.add_argument('--after', required=True)
parser.add_argument('--node', required=True)
parser.add_argument('--bun', required=True)
parser.add_argument('--source', required=True)
parser.add_argument('--output', required=True)
args = parser.parse_args()
commands = {
    'before': [args.before],
    'after': [args.after],
    'Node': [args.node, args.source],
    'Bun': [args.bun, args.source],
}
expected = b'-0.169075164\n-0.169086185\n'
environment = os.environ.copy()
for variable in ('NODE_OPTIONS', 'BUN_OPTIONS', 'ADAMIC_THREADS'):
    environment.pop(variable, None)

def execute(command):
    result = subprocess.run(command, env=environment, capture_output=True,
                            check=True, timeout=30)
    if result.stdout != expected:
        raise RuntimeError(f'checksum mismatch: {command}: {result.stdout!r}')

for command in commands.values():
    execute(command)
report = {
    'machine': platform.node(),
    'kernel': platform.platform(),
    'affinity': sorted(os.sched_getaffinity(0)),
    'cpu_max': open('/sys/fs/cgroup/cpu.max').read().strip(),
    'commands': commands,
    'load_before': os.getloadavg(),
    'samples': {name: [] for name in commands},
}
names = list(commands)
for round_index in range(5):
    first = round_index % len(names)
    for name in names[first:] + names[:first]:
        user_before = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime
        start = time.perf_counter()
        execute(commands[name])
        wall = time.perf_counter() - start
        user = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime - user_before
        report['samples'][name].append({'wall': wall, 'user': user})
report['load_after'] = os.getloadavg()
with open(args.output, 'w') as output:
    json.dump(report, output, indent=2)
    output.write('\n')
print(json.dumps({name: min(samples, key=lambda sample: sample['wall'])
                  for name, samples in report['samples'].items()}, indent=2))
