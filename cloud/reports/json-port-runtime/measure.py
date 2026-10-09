#!/usr/bin/env python3
"""Five interleaved full-corpus before/after runs; exact Go output every time."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import resource
import subprocess
import time

p = argparse.ArgumentParser()
p.add_argument('directory', type=Path)
p.add_argument('output', type=Path)
a = p.parse_args()
d = a.directory.resolve()
expected = (d / 'expected.txt').read_bytes()
input_file = d / 'cases.txt'
env = os.environ.copy()
for name in ('ADAMIC_THREADS', 'NODE_OPTIONS', 'BUN_OPTIONS'):
    env.pop(name, None)
commands = {name: [str(d / name), '--cases', str(input_file)] for name in ('before', 'after')}

# The standalone Go driver must agree with the test helpers' saved answer stream.
def execute(name, command):
    out, err = d / (name + '.timed.out'), d / (name + '.timed.err')
    user = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime
    with out.open('wb') as stdout, err.open('wb') as stderr:
        start = time.perf_counter()
        subprocess.run(command, env=env, stdout=stdout, stderr=stderr, check=True, timeout=300)
        wall = time.perf_counter() - start
    user = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime - user
    if err.stat().st_size or out.read_bytes() != expected:
        raise RuntimeError(name + ' differs from Go or writes stderr')
    return {'wall': wall, 'user': user}

execute('Go-validation', [str(d / 'go-port'), '--cases', str(input_file)])
r = {
    'machine': platform.node(), 'platform': platform.platform(),
    'affinity': sorted(os.sched_getaffinity(0)),
    'cpu_max': Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
    'input_sha256': hashlib.sha256(input_file.read_bytes()).hexdigest(),
    'output_sha256': hashlib.sha256(expected).hexdigest(),
    'input_bytes': input_file.stat().st_size, 'output_bytes': len(expected),
    'cases': input_file.read_bytes().count(b'\n'),
    'commands': commands, 'load_before': os.getloadavg(),
    'samples': {name: [] for name in commands}, 'round_loads': [],
}
for round_index in range(5):
    names = list(commands) if round_index % 2 == 0 else list(reversed(commands))
    for name in names:
        result = execute(name, commands[name])
        r['samples'][name].append(result)
        print(f'round {round_index + 1} {name}: {result}; exact Go bytes', flush=True)
    r['round_loads'].append(os.getloadavg())
r['load_after'] = os.getloadavg()
a.output.write_text(json.dumps(r, indent=2) + '\n')
print(json.dumps({name: min(values, key=lambda x: x['wall']) for name, values in r['samples'].items()}, indent=2))
