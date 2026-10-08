#!/usr/bin/env python3
"""Check that compiler comparison validation rejects independent evidence mutations."""
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

here = Path(__file__).resolve().parent
root = here / 'evidence/compiler-tips'
with tempfile.TemporaryDirectory(prefix='tsc-entry-compiler-mutants-') as scratch:
    for name in ('compiler-pin', 'binary-provenance', 'source-hash', 'stop-population', 'split-byte', 'node-byte', 'classification', 'first-stop', 'probe-diagnostic'):
        directory = Path(scratch) / name
        shutil.copytree(root, directory)
        target = directory / 'area-b68b2fe1'
        p = target / 'comparison.json'
        report = json.loads(p.read_text())
        if name == 'compiler-pin':
            report['compiler_commit'] = '0' * 40
        elif name == 'binary-provenance':
            binary = target / 'binary.log'
            binary.write_text(binary.read_text().replace('vcs.modified=false', 'vcs.modified=true'))
        elif name == 'source-hash':
            report['source_hashes']['src/tsc/tsc.ts'] = '0' * 64
        elif name == 'stop-population':
            report['stops'].pop()
        elif name == 'classification':
            report['stops'][0]['status'] = 'disappear' if report['stops'][0]['status'] == 'remain' else 'remain'
        elif name == 'first-stop':
            report['first_stop'] = 'changed'
        elif name == 'probe-diagnostic':
            probe = target / '01-probe-types.stderr'
            probe.write_text('changed\n')
        elif name == 'split-byte':
            stream = target / 'build-1.stderr'
            stream.write_bytes(stream.read_bytes() + b'changed\n')
        else:
            (target / '01-probe-node.stdout').write_text('changed\n')
        p.write_text(json.dumps(report))
        with (root / f'mutant-{name}.log').open('wb') as output:
            result = subprocess.run(['python3', str(here / 'verify-compilers.py'), str(directory)], stdout=output, stderr=subprocess.STDOUT)
        if result.returncode != 1:
            raise RuntimeError(f'{name} survived: exit {result.returncode}')
        print(f'{name}: caught, exit 1')
