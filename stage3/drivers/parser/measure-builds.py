#!/usr/bin/env python3
"""Retain actual split/off/repeated attempts; never label pre-clang time as clang."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time

p = argparse.ArgumentParser()
p.add_argument('compiler', type=Path)
p.add_argument('metrics', type=Path)
p.add_argument('entry', type=Path)
p.add_argument('scratch', type=Path)
p.add_argument('output', type=Path)
p.add_argument('--jobs', type=int, required=True)
a = p.parse_args()
for name in ['compiler', 'metrics', 'entry', 'scratch', 'output']:
    setattr(a, name, getattr(a, name).resolve())
a.output.mkdir(exist_ok=False)
cache = a.output / 'cache'
cache.mkdir()
result = {'jobs': a.jobs, 'cache': str(cache), 'generated_c': None, 'attempts': []}
c = a.output / 'parser.c'
with c.open('wb') as stdout, (a.output / 'emit.stderr').open('wb') as stderr:
    emitted = subprocess.run([str(a.compiler), 'c', str(a.entry)], cwd=a.scratch, stdout=stdout, stderr=stderr)
result['emit_exit'] = emitted.returncode
if emitted.returncode == 0:
    data = c.read_bytes()
    result['generated_c'] = {'bytes': len(data), 'lines': len(data.splitlines())}
for mode in ['unsplit', 'split', 'split-repeat']:
    binary = a.output / ('parser-' + mode)
    env = dict(os.environ, XDG_CACHE_HOME=str(cache),
               ADAMIC_NATIVE_SPLIT='0' if mode == 'unsplit' else '1',
               ADAMIC_NATIVE_JOBS=str(a.jobs))
    env.pop('ADAMIC_GATE_UNCACHED', None)
    if emitted.returncode == 0:
        command = [str(a.metrics), str(c), str(binary), mode, str(a.jobs), str(a.output / (mode + '-native.json'))]
    else:
        command = [str(a.compiler), 'build', str(a.entry), '-o', str(binary)]
    started = time.monotonic()
    with (a.output / (mode + '.stdout')).open('wb') as stdout, (a.output / (mode + '.stderr')).open('wb') as stderr:
        built = subprocess.run(command, cwd=a.scratch, env=env, stdout=stdout, stderr=stderr)
    row = {'mode': mode, 'exit': built.returncode,
           'attempt_wall_seconds': time.monotonic() - started,
           'generated_c': result['generated_c'], 'clang_wall_seconds': None,
           'binary_bytes': binary.stat().st_size if binary.exists() else None}
    measured = a.output / (mode + '-native.json')
    if measured.exists():
        row['native_build_phase'] = json.loads(measured.read_text())
    else:
        row['clang_status'] = 'unreached: checking/lowering failed before C emission'
    row['cache_files_after'] = sum(path.is_file() for path in cache.rglob('*'))
    result['attempts'].append(row)
result['split_repeat_is_warm'] = emitted.returncode == 0 and result['attempts'][1]['exit'] == 0
(a.output / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2))
