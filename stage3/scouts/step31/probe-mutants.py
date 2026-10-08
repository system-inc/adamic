#!/usr/bin/env python3
"""Prove the three API seam assertions reject independent behavior changes."""
import argparse
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('output', type=Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=False)
source = (ROOT / 'piece-probes.cjs').read_text()
mutants = [
    ('checker-input', '"wrong";', '42;', 'diagnostics.map'),
    ('printer-comments', 'newLine: ts.NewLineKind.LineFeed}', 'newLine: ts.NewLineKind.LineFeed, removeComments: true}', 'assert.equal(printed'),
    ('emitter-transform', 'ts.getTransformers(options)', '{scriptTransformers: [], declarationTransformers: []}', 'assert.deepEqual(emitted'),
]
for name, before, after, assertion in mutants:
    if before not in source:
        raise RuntimeError('missing mutant anchor')
    mutant = args.output / (name + '.cjs')
    mutant.write_text(source.replace(before, after, 1))
    log = args.output / (name + '.log')
    with log.open('wb') as stream:
        result = subprocess.run(['node', str(mutant)], stdout=stream, stderr=stream, timeout=120)
    text = log.read_text()
    line = next(index for index, value in enumerate(source.splitlines(), 1) if assertion in value)
    if result.returncode != 1 or 'AssertionError' not in text or f'{mutant}:{line}:' not in text:
        raise RuntimeError(name + ': expected seam assertion was not the catcher; see ' + str(log))
    print(name + ': caught by ' + assertion + ', exit 1')
