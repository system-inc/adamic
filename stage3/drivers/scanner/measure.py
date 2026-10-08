#!/usr/bin/env python3
"""Measure a green scanner proof on its identical recorded corpus, best of three."""
import argparse
import json
import os
from pathlib import Path
import resource
import shutil
import subprocess
import sys

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('run', type=Path, help='green run.sh output directory')
args = parser.parse_args()
run = args.run.resolve()
report = json.loads((run / 'report.json').read_text())
if report.get('native_diff_exit') != 0 or report.get('native_exit') != 0:
    sys.exit('measurement requires a green native comparison')
repository = Path(__file__).resolve().parents[3]
cache = Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3')))
env = dict(os.environ, SCANNER_TYPESCRIPT=str(cache / 'api/node_modules/typescript/lib/typescript.js'),
           SCANNER_RUNTIME=str(repository / 'oracle/adamic.mjs'))
inputs = Path(report['input_tree'])
paths = [str(inputs / path) for path in json.loads((run / 'files.json').read_text())]
commands = {
    'native': [str(run / 'scanner-native'), *paths],
    'node': ['node', '--disable-warning=ExperimentalWarning',
             str(Path(__file__).with_name('node.mjs')), str(run / 'main.a'), *paths],
}
output = run / 'measurement'
output.mkdir()
result = {'native_binary_bytes': (run / 'scanner-native').stat().st_size,
          'method': 'RUSAGE_CHILDREN user CPU seconds delta, sequential best of three; includes process startup and output',
          'files': len(paths)}
reference = run / 'node.stdout'
for name, command in commands.items():
    times = []
    for iteration in range(1, 4):
        stdout = output / f'{name}-{iteration}.stdout'
        with stdout.open('wb') as out, (output / f'{name}-{iteration}.stderr').open('wb') as err:
            before = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime
            code = subprocess.run(command, cwd=run, env=env, stdout=out, stderr=err).returncode
            user = resource.getrusage(resource.RUSAGE_CHILDREN).ru_utime - before
        with (output / f'{name}-{iteration}.diff').open('wb') as log:
            diff = subprocess.run(['diff', '-u', str(reference), str(stdout)], stdout=log, stderr=log).returncode
        if code or diff:
            sys.exit(f'{name} measurement {iteration} failed: exit {code}, diff {diff}')
        times.append(user)
    result[name] = {'user_seconds': times, 'best_user_seconds': min(times)}
perf = shutil.which('perf')
if not perf:
    result['instructions'] = 'unavailable: perf is not installed'
else:
    result['instructions'] = {}
    for name, command in commands.items():
        count = output / f'{name}-perf.txt'
        stdout = output / f'{name}-perf.stdout'
        with stdout.open('wb') as out, (output / f'{name}-perf.stderr').open('wb') as err:
            code = subprocess.run([perf, 'stat', '-x', ',', '-e', 'instructions', '-o', str(count), '--', *command],
                                  cwd=run, env=env, stdout=out, stderr=err).returncode
        observation = {'exit': code, 'stat': count.read_text() if count.exists() else '',
                       'stderr': (output / f'{name}-perf.stderr').read_text()}
        if code == 0:
            with (output / f'{name}-perf.diff').open('wb') as log:
                observation['diff_exit'] = subprocess.run(['diff', '-u', str(reference), str(stdout)], stdout=log, stderr=log).returncode
            if observation['diff_exit']:
                sys.exit(f'{name} perf output differs')
        result['instructions'][name] = observation
(output / 'report.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result))
