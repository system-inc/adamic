#!/usr/bin/env python3
"""Measure directory-cold and warm installs in alternating order, with full build flags."""
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time

repository = Path(__file__).resolve().parent.parent
output = Path(sys.argv[1])
output.mkdir(parents=True, exist_ok=True)
scratch = Path(tempfile.mkdtemp(prefix='markdown-timings-', dir='/tmp/adamic-gate'))
node = Path(os.environ.get('ADAMIC_TOOLS', '/opt/adamic-tools')) / 'bin/node'
helper = repository / 'cloud/setup-markdown-width.py'
source = repository / 'cloud/markdown-width'


def command(*arguments):
    return subprocess.check_output(arguments, cwd=repository).decode().strip()


def flags():
    return dict(commit=command('git', 'rev-parse', 'HEAD'), nproc=command('nproc'),
                cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                go=command('go', 'version'), clang=command('clang', '--version').splitlines()[0],
                node=command(str(node), '--version'), load=Path('/proc/loadavg').read_text().strip())


results = []
for loop in range(1, 4):
    cold = scratch / f'cold-{loop}'
    warm = cold if loop != 2 else scratch / 'cold-1'
    stages = [('cold', cold), ('warm', warm)] if loop != 2 else [('warm', warm), ('cold', cold)]
    for stage, destination in stages:
        arguments = ['python3', str(helper), str(source), str(destination), str(node)]
        before = flags()
        started = time.monotonic()
        log = output / f'{stage}-{loop}.log'
        with log.open('wb') as stream:
            result = subprocess.run(arguments, stdout=stream, stderr=subprocess.STDOUT,
                                    env=dict(os.environ, ADAMIC_GATE_UNCACHED='0'))
        elapsed = time.monotonic() - started
        after = flags()
        answer = log.read_text()
        assert result.returncode == 0, answer
        assert ('installed (npm ci' if stage == 'cold' else 'skipped (validated lock') in answer, answer
        record = dict(loop=loop, stage=stage, seconds=elapsed, command=arguments,
                      before=before, after=after, cached=stage == 'warm', exit=result.returncode)
        results.append(record)
        with log.open('a') as stream:
            stream.write('build-flags ' + json.dumps(record, sort_keys=True) + '\n')
        (output / 'timings.json').write_text(json.dumps(results, indent=2) + '\n')
        print(stage, loop, f'{elapsed:.6f}s', flush=True)
