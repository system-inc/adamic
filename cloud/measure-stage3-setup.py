#!/usr/bin/env python3
"""Measure fresh API installs and byte-validated warm hits in alternating order."""
from pathlib import Path
import json
import os
import shutil
import subprocess
import sys
import tempfile
import time

repository = Path(__file__).resolve().parent.parent
output = Path(sys.argv[1]); output.mkdir(parents=True, exist_ok=True)
scratch = Path(tempfile.mkdtemp(prefix='stage3-timings-', dir='/tmp/adamic-gate'))
node = Path(os.environ.get('ADAMIC_TOOLS', '/opt/adamic-tools')) / 'bin/node'


def command(*args):
    return subprocess.check_output(args, cwd=repository).decode().strip()


def flags():
    return dict(commit=command('git', 'rev-parse', 'HEAD'), nproc=command('nproc'),
                cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(), go=command('go', 'version'),
                clang=command('clang', '--version').splitlines()[0], node=command(str(node), '--version'),
                load=Path('/proc/loadavg').read_text().strip())


results = []
for loop in range(1, 4):
    cold = scratch / f'cold-{loop}'
    shutil.copytree(repository / 'cloud/testdata/stage3-api', cold / 'stage3/api')
    warm = cold if loop != 2 else scratch / 'cold-1'
    order = [('cold', cold), ('warm', warm)] if loop != 2 else [('warm', warm), ('cold', cold)]
    for mode, seat in order:
        args = ['python3', str(repository / 'cloud/setup-stage3-api.py'), str(seat), str(seat / 'tools'), str(node)]
        before = flags(); started = time.monotonic()
        log = output / f'{mode}-{loop}.log'
        with log.open('wb') as stream:
            result = subprocess.run(args, env=dict(os.environ, ADAMIC_GATE_UNCACHED='0'),
                                    stdout=stream, stderr=subprocess.STDOUT)
        elapsed = time.monotonic() - started; after = flags()
        answer = log.read_text()
        assert result.returncode == 0 and ('installed (npm ci' if mode == 'cold' else 'skipped (validated API lock') in answer, answer
        record = dict(loop=loop, mode=mode, seconds=elapsed, command=args, before=before, after=after,
                      cached=mode == 'warm', exit=result.returncode)
        results.append(record)
        (output / 'timings.json').write_text(json.dumps(results, indent=2) + '\n')
        with log.open('a') as stream:stream.write('build-flags ' + json.dumps(record) + '\n')
        print(loop, mode, f'{elapsed:.6f}s', flush=True)
