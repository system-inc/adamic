#!/usr/bin/env python3
"""Prove each byte/process comparison can fail independently."""
import argparse
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('output', type=Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=False)
golden = ROOT / 'fixtures/golden.jsonl'
source = golden.read_bytes()
for name, stdout, stderr, exit_code, failures in [
    ('baseline', source, b'', 0, []),
    ('stdout-byte', source.replace(b'"flags":', b'"Flags":', 1), b'', 0, ['stdout']),
    ('stderr-byte', source, b'x', 0, ['stderr']),
    ('exit-byte', source, b'', 1, ['exit']),
]:
    fixture = args.output / (name + '.stdout')
    fixture.write_bytes(stdout)
    script = 'import pathlib,sys;sys.stdout.buffer.write(pathlib.Path(sys.argv[1]).read_bytes());sys.stderr.buffer.write(sys.argv[2].encode());sys.exit(int(sys.argv[3]))'
    directory = args.output / name
    with (args.output / (name + '.log')).open('wb') as log:
        result = subprocess.run([sys.executable, str(ROOT / 'compare.py'), '--output', str(directory), str(golden), '--',
            sys.executable, '-c', script, str(fixture), stderr.decode(), str(exit_code)], stdout=log, stderr=log)
    report = json.loads((directory / 'report.json').read_text())
    assert result.returncode == bool(failures), (name, result.returncode)
    assert sorted(report['differences']) == failures, (name, report)
    if name == 'stdout-byte':
        assert len(stdout) == len(source) and sum(a != b for a, b in zip(stdout, source)) == 1
    print(f'{name}: exit {result.returncode}, differences {failures}')
