"""Isolated alternating release runs, one load and all six rules per process.

Usage: bench.py directory before-executable after-executable Go-oracle config manifest
"""
from pathlib import Path
import os
import subprocess
import sys
import time

directory = Path(sys.argv[1])
directory.mkdir(parents=True, exist_ok=True)
binaries = dict(zip(('before', 'after', 'oracle'), sys.argv[2:5]))
config, manifest = sys.argv[5:7]
environment = dict(os.environ)
environment['ADAMIC_TSGO_TIMING'] = '1'
assert 'ADAMIC_TSGO_PROFILE' not in environment, 'Throughput runs must not profile'
for round_, order in enumerate((('before', 'after', 'oracle'), ('oracle', 'after', 'before'), ('after', 'before', 'oracle')), 1):
    for name in order:
        command = [binaries[name], config, manifest, '--count']
        stdout = directory / f'bench-{round_}-{name}.stdout.log'
        stderr = directory / f'bench-{round_}-{name}.stderr.log'
        with stdout.open('wb') as out, stderr.open('wb') as err:
            start = time.monotonic()
            result = subprocess.run(command, stdout=out, stderr=err, env=environment)
            elapsed = time.monotonic() - start
        assert result.returncode == 0, (name, result.returncode)
        assert stdout.read_bytes() == b'findings 1763\n'
        print(f'round={round_} name={name} process_s={elapsed:.9f} {stderr.read_text().strip()}', flush=True)
