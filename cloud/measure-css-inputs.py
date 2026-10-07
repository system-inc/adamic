#!/usr/bin/env python3
"""Measure just the added gate work, interleaving cold and warm on one box."""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time

cloud = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('gate', cloud / 'setup-gate-inputs.py')
gate = importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
report = cloud / 'reports/css-gate-inputs'
root = Path(os.environ['ADAMIC_TOOLS']) / 'gate-inputs'
node = str(Path(os.environ['ADAMIC_TOOLS']) / 'bin/node')
records = []

def flags():
    def output(args):return subprocess.check_output(args).decode().strip()
    return dict(commit=output(['git', 'rev-parse', 'HEAD']), nproc=output(['nproc']),
                cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                go=output(['go', 'version']), clang=output(['clang', '--version']).splitlines()[0],
                node=output([node, '--version']), GOFLAGS=os.environ.get('GOFLAGS', ''),
                load_before=Path('/proc/loadavg').read_text().strip())

with tempfile.TemporaryDirectory(prefix='css-input-cost-', dir='/workspace') as temporary:
    for loop in range(1, 4):
        inputs = Path(temporary) / str(loop);inputs.mkdir()
        for mode in ['cold', 'warm']:
            build_flags = flags();build_flags['cached'] = mode == 'warm';start = time.monotonic()
            with (report / f'cost-{mode}-{loop}.log').open('w') as log:
                result = subprocess.run([sys.executable, '-c',
                    'import importlib.util, pathlib, sys; s=importlib.util.spec_from_file_location("gate",sys.argv[1]);g=importlib.util.module_from_spec(s);s.loader.exec_module(g);print(g.css_fixtures(pathlib.Path(sys.argv[2])))',
                    str(cloud / 'setup-gate-inputs.py'), str(inputs)], stdout=log, stderr=subprocess.STDOUT)
            seconds = time.monotonic() - start
            assert result.returncode == 0
            build_flags['load_after'] = Path('/proc/loadavg').read_text().strip()
            records.append(dict(loop=loop, step='CSS fixtures', mode=mode, seconds=seconds, build_flags=build_flags,
                                instrument=f'python3 cloud/measure-css-inputs.py: css_fixtures({inputs})'))
            build_flags = flags();build_flags['cached'] = True;start = time.monotonic()
            answer = gate.shared_prettier(root, node);seconds = time.monotonic() - start
            build_flags['load_after'] = Path('/proc/loadavg').read_text().strip()
            records.append(dict(loop=loop, step='shared Prettier paths', mode=mode, seconds=seconds, answer=answer,
                                build_flags=build_flags, instrument=f'python3 cloud/measure-css-inputs.py: shared_prettier({root}, {node})'))
            print(mode, loop, records[-2]['seconds'], seconds, flush=True)
    original = gate.artifact_digest(inputs / 'css-fixtures')
    with (report / 'uncached-fixtures.log').open('w') as log:
        result = subprocess.run([sys.executable, '-c',
            'import importlib.util,pathlib,sys;s=importlib.util.spec_from_file_location("g",sys.argv[1]);g=importlib.util.module_from_spec(s);s.loader.exec_module(g);print(g.css_fixtures(pathlib.Path(sys.argv[2])))',
            str(cloud / 'setup-gate-inputs.py'), str(inputs)], env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=log, stderr=subprocess.STDOUT)
    assert result.returncode == 0 and gate.artifact_digest(inputs / 'css-fixtures') == original
    records.append(dict(proof='uncached fixture bytes and modes identical', digest=original))
(report / 'costs.json').write_text(json.dumps(records, indent=2) + '\n')
