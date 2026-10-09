#!/usr/bin/env python3
"""Time complete YAML layer drivers; every measured output must equal Go's bytes."""
import argparse
from pathlib import Path
import resource
import statistics
import subprocess
import time

# An inherited CPU-time hang guard does not charge time spent waiting for a loaded box.
# These drivers may buffer until exit, so allow thirty minutes of CPU before stopping.
def cpu_hang_guard():
    resource.setrlimit(resource.RLIMIT_CPU, (30 * 60, 30 * 60 + 1))


parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('artifacts', type=Path)
parser.add_argument('library', type=Path)
parser.add_argument('--layer', choices=['lexer', 'scalar', 'format'], default='lexer')
parser.add_argument('--rounds', type=int, default=5)
args = parser.parse_args()
artifacts = args.artifacts.resolve()
root = Path(__file__).resolve().parents[3]
layer = args.layer
cases = str(artifacts / f'{layer}-cases.txt')
expected = (artifacts / f'{layer}-expected.txt').read_bytes()
count = (artifacts / f'{layer}-cases.txt').read_bytes().count(b'\n')
driver = 'main' if layer == 'format' else ('lex_main' if layer == 'lexer' else 'scalar_main')
case_args = ['--cases', cases] if layer == 'format' else [cases]
commands = {
    'native': [str(artifacts / f'native-{layer}'), *case_args],
    'Node source': ['node', '--disable-warning=ExperimentalWarning', str(root / 'oracle/node.mjs'), str(root / f'stage1/cohere/yaml/{driver}.ts'), *case_args],
    'Go': [str(artifacts / f'go-{layer}'), *case_args],
    'yaml 2.9.0': ['node', str(root / f'stage1/cohere/yaml/testdata/{layer}_library.mjs'), str(args.library.resolve()), cases],
}
if layer == 'format':
    del commands['yaml 2.9.0']  # Published Prettier differs on the 42 proved upstream cases.
samples = {name: [] for name in commands}
for round_number in range(args.rounds):
    for name, command in commands.items():
        stdout = artifacts / (name + '.timed.stdout')
        stderr = artifacts / (name + '.timed.stderr')
        with stdout.open('wb') as out, stderr.open('wb') as err:
            start = time.perf_counter()
            subprocess.run(command, stdout=out, stderr=err, check=True, preexec_fn=cpu_hang_guard)
            elapsed = time.perf_counter() - start
        if stdout.read_bytes() != expected or stderr.stat().st_size:
            raise RuntimeError(f'{name}: bytes or stderr differ from Go')
        samples[name].append(elapsed)
        print(f'round {round_number + 1} {name}: {elapsed:.6f}s, {count / elapsed:.2f} texts/s; exact Go bytes', flush=True)
for name, times in samples.items():
    print(f'{name}: median {count / statistics.median(times):.2f} texts/s; seconds {times}')
print(f'{count} {layer} cases; {len(expected)} answer bytes; startup, input and output included')
