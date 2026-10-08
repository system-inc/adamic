#!/usr/bin/env python3
"""Run source Node, sanitized native, release native, and backend Node.
Run from repository root with cloud/setup.sh's environment sourced.
All command output is saved without piping. Build failures are observations.
"""
import argparse
import json
import os
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--compiler', default='/tmp/nodehost-adamic')
parser.add_argument('--logs', default='/tmp/nodehost-runs')
parser.add_argument('--io-shim', help='Linux shared library injected into executions only')
parser.add_argument('programs', nargs='*')
args = parser.parse_args()
root = Path.cwd()
unit = root / 'notes/coverage-oct8/nodehost'
logs = Path(args.logs).resolve()
logs.mkdir(parents=True, exist_ok=True)
env = dict(os.environ, ASAN_OPTIONS='detect_leaks=1:halt_on_error=1', UBSAN_OPTIONS='halt_on_error=1')
def command(name, argv):
    with (logs / (name + '.stdout')).open('wb') as out, (logs / (name + '.stderr')).open('wb') as err:
        try:
            run_env = dict(env)
            if args.io_shim and not name.endswith('.build'):
                run_env['LD_PRELOAD'] = str(Path(args.io_shim).resolve())
            result = subprocess.run(argv, cwd=root, env=run_env, stdout=out, stderr=err, timeout=180)
            status = result.returncode
        except subprocess.TimeoutExpired:
            status = 124
    (logs / (name + '.exit')).write_text(str(status) + '\n')
    (logs / (name + '.command.json')).write_text(json.dumps(argv) + '\n')
    return {'exit': status, 'stdout': (logs / (name + '.stdout')).read_text(errors='backslashreplace'), 'stderr': (logs / (name + '.stderr')).read_text(errors='backslashreplace')}
results = {}
programs = [unit / name for name in args.programs] if args.programs else sorted(unit.glob('*.a'))
for path in programs:
    name = path.stem
    rows = {}
    rows['node'] = command(name + '.node', ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(path)])
    for mode in ['sanitize', 'release']:
        binary = logs / (name + '.' + mode + '.bin')
        build = command(name + '.' + mode + '.build', [args.compiler, 'build', str(path), '-o', str(binary)] + (['--sanitize'] if mode == 'sanitize' else []))
        rows[mode] = {'build': build}
        if build['exit'] == 0:
            rows[mode]['run'] = command(name + '.' + mode, [str(binary)])
    build = command(name + '.javascript.build', [args.compiler, 'js', str(path)])
    rows['javascript'] = {'build': {'exit': build['exit'], 'stderr': build['stderr']}}
    if build['exit'] == 0:
        script = logs / (name + '.mjs')
        script.write_text(build['stdout'])
        rows['javascript']['run'] = command(name + '.javascript', ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(script)])
    actual = [rows['node']] + [rows[mode].get('run') for mode in ['sanitize', 'release', 'javascript']]
    # Equality uses raw bytes from files, not decoded JSON strings.
    modes = ['node', 'sanitize', 'release', 'javascript']
    complete = all(row is not None for row in actual)
    rows['agree'] = complete and all((logs / (name + '.' + mode + '.' + part)).read_bytes() == (logs / (name + '.node.' + part)).read_bytes() for mode in modes[1:] for part in ['stdout', 'stderr', 'exit'])
    results[name] = rows
    print(name, 'AGREE' if rows['agree'] else 'DIFFER', flush=True)
(logs / 'results.json').write_text(json.dumps(results, indent=2, ensure_ascii=True) + '\n')
