#!/usr/bin/env python3
"""Compare the same dense source on two checkouts, then measure alternating runs.
Source the pinned toolchain environment and set GOPROXY before running this.
"""
import argparse
import hashlib
from pathlib import Path
import statistics
import subprocess
import tempfile
import time

parser = argparse.ArgumentParser()
parser.add_argument('--before', required=True, type=Path)
parser.add_argument('--after', required=True, type=Path)
parser.add_argument('--rounds', type=int, default=17)
args = parser.parse_args()
source = args.after.resolve() / 'internal/native/performance/array-holes/hot-loop.a'
expected = b'680000000\n'
truth = subprocess.run(['node', '--disable-warning=ExperimentalWarning', str(args.after.resolve() / 'oracle/node.mjs'), str(source)], capture_output=True, check=True)
assert truth.stdout == expected and not truth.stderr, truth
with tempfile.TemporaryDirectory(prefix='adamic-array-holes-measure-') as temp:
    code = {}
    binaries = {}
    for name, root in [('before', args.before.resolve()), ('after', args.after.resolve())]:
        binary = str(Path(temp) / name)
        subprocess.run(['go', 'run', './cmd/adamic', 'build', str(source), '-o', binary], cwd=root, check=True)
        code[name] = subprocess.run(['go', 'run', './cmd/adamic', 'c', str(source)], cwd=root, capture_output=True, check=True).stdout
        binaries[name] = binary
    print('generated C identical:', code['before'] == code['after'])
    print('generated C sha256:', hashlib.sha256(code['after']).hexdigest())
    samples = {'before': [], 'after': []}
    for round_ in range(args.rounds):
        for name in (['before', 'after'] if round_ % 2 == 0 else ['after', 'before']):
            start = time.perf_counter()
            result = subprocess.run([binaries[name]], capture_output=True, check=True)
            elapsed = time.perf_counter() - start
            assert result.stdout == expected and not result.stderr, result
            if round_ > 1:
                samples[name].append(elapsed)
    for name, values in samples.items():
        print(name, 'seconds:', values, 'median:', statistics.median(values))
    print('after/before median:', statistics.median(samples['after']) / statistics.median(samples['before']))
