#!/usr/bin/env python3
"""Prove the evidence checks can reject independent artifact mutations."""
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

here = Path(__file__).resolve().parent
with tempfile.TemporaryDirectory(prefix='tsc-entry-mutants-') as scratch:
    for name in ('probe-diagnostic', 'node-byte', 'split-byte', 'stop-population'):
        directory = Path(scratch) / name
        shutil.copytree(here / 'evidence', directory)
        if name == 'probe-diagnostic':
            shutil.copyfile(directory / '01-defined-control.log', directory / '01-indexed-path-build.stderr')
        elif name == 'node-byte':
            (directory / '01-indexed-path-node.stdout').write_text('changed\n')
        elif name == 'split-byte':
            p = directory / '01-split-1.stderr'
            p.write_bytes(p.read_bytes() + b'changed\n')
        else:
            p = directory / 'stops.json'
            rows = json.loads(p.read_text())
            p.write_text(json.dumps(rows[:-1]))
        log = here / 'evidence' / f'mutant-{name}.log'
        with log.open('wb') as output:
            result = subprocess.run(['python3', str(here / 'verify.py'), str(directory)], stdout=output, stderr=subprocess.STDOUT)
        if result.returncode != 1:
            raise RuntimeError(f'{name} survived: exit {result.returncode}')
        print(f'{name}: caught, exit 1; {log.read_text().strip()}')
