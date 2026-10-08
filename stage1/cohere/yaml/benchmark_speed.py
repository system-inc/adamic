#!/usr/bin/env python3
"""Interleave before/after YAML drivers and verify every timed byte against Go."""
import argparse
import hashlib
import json
from pathlib import Path
import statistics
import subprocess
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('artifacts', type=Path)
parser.add_argument('snapshots', type=Path)
parser.add_argument('--rounds', type=int, default=5)
args = parser.parse_args()
artifacts = args.artifacts.resolve()
snapshots = args.snapshots.resolve()
root = Path(__file__).resolve().parents[3]
cases = artifacts / 'format-cases.txt'
expected = (artifacts / 'format-expected.txt').read_bytes()
count = cases.read_bytes().count(b'\n')
inputs = ['--cases', str(cases)]
node = ['node', '--disable-warning=ExperimentalWarning', str(root / 'oracle/node.mjs')]
commands = {
    'baseline native': [str(snapshots / 'baseline'), *inputs],
    'shared schemas': [str(snapshots / 'shared-schema'), *inputs],
    'fast paths with Map': [str(snapshots / 'fast-path-maps'), *inputs],
    'native': [str(snapshots / 'final'), *inputs],
    'baseline Node': [*node, str(snapshots / 'baseline-source/main.ts'), *inputs],
    'Node': [*node, str(root / 'stage1/cohere/yaml/main.ts'), *inputs],
    'Go': [str(artifacts / 'go-format'), *inputs],
}
samples = {name: [] for name in commands}
loads = []
names = list(commands)
for round_number in range(args.rounds):
    order = names[round_number % len(names):] + names[:round_number % len(names)]
    for name in order:
        stdout = snapshots / (name + '.timed.stdout')
        stderr = snapshots / (name + '.timed.stderr')
        with stdout.open('wb') as out, stderr.open('wb') as err:
            started = time.perf_counter()
            subprocess.run(commands[name], stdout=out, stderr=err, check=True)
            elapsed = time.perf_counter() - started
        if stdout.read_bytes() != expected or stderr.stat().st_size:
            raise RuntimeError(f'{name}: bytes or stderr differ from Go')
        samples[name].append(elapsed)
        print(f'round {round_number + 1} {name}: {elapsed:.6f}s, {count / elapsed:.2f} texts/s; exact Go bytes', flush=True)
    loads.append(Path('/proc/loadavg').read_text().strip())
result = dict(texts=count, answer_bytes=len(expected),
              input_sha256=hashlib.sha256(cases.read_bytes()).hexdigest(),
              answer_sha256=hashlib.sha256(expected).hexdigest(),
              rounds=samples,
              median_texts_per_second={name: count / statistics.median(values) for name, values in samples.items()},
              loadavg=loads)
(snapshots / 'measurements.txt').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2))
