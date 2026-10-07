#!/usr/bin/env python3
"""Prove the instruction guard catches selecting the DFA for short VM calls."""
import argparse
import pathlib
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('baseline_instructions', type=pathlib.Path)
parser.add_argument('output', type=pathlib.Path)
args = parser.parse_args()
root = pathlib.Path(__file__).resolve().parents[2]
source = root / 'internal/native/runtime/regexp.c'
original = source.read_text()
old, new = 'ascii_length >= 2048', 'ascii_length < 2048'
if original.count(old) != 1:
    raise SystemExit('wrong mutation site count')
args.output.mkdir(parents=True, exist_ok=True)
try:
    source.write_text(original.replace(old, new))
    with (args.output / 'build.log').open('w') as log:
        subprocess.run(['go', 'run', './bench/regex', '-out', str(args.output),
                        '-build-only'], cwd=root, stdout=log, stderr=log, check=True)
    with (args.output / 'instructions.csv').open('w') as counts, \
         (args.output / 'count.log').open('w') as log:
        subprocess.run(['python3', 'bench/regex/count.py', str(args.output)],
                       cwd=root, stdout=counts, stderr=log, check=True)
    with (args.output / 'guard.log').open('w') as log:
        result = subprocess.run(['python3', 'bench/regex/check-dispatch.py',
                                 str(args.baseline_instructions),
                                 str(args.output / 'instructions.csv')],
                                cwd=root, stdout=log, stderr=log)
    if result.returncode == 0 or 'INSTRUCTION REGRESSION' not in (args.output / 'guard.log').read_text():
        raise SystemExit('short-input DFA mutant NOT CAUGHT by instruction guard')
    print(f'short-input DFA mutant caught by instruction guard: {args.output / "guard.log"}')
finally:
    source.write_text(original)
