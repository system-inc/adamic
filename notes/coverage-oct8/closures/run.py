#!/usr/bin/env python3
"""Capture each stage separately; compare execution bytes only after compilation succeeds."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

parser = argparse.ArgumentParser()
parser.add_argument('--compiler', default='/tmp/coverage-adamic')
parser.add_argument('--results', default='notes/coverage-oct8/closures/results')
parser.add_argument('names', nargs='*')
args = parser.parse_args()
root = Path(__file__).resolve().parents[3]
unit = Path(__file__).resolve().parent
results = Path(args.results).resolve()
results.mkdir(parents=True, exist_ok=True)
work = Path('/tmp/coverage-closures') / hashlib.sha256(str(results).encode()).hexdigest()[:12]
work.mkdir(parents=True, exist_ok=True)
records = {}
expected_refusals = {
    'delayed_fill_guard': 'delayed fill of an array with holes',
    'mixed_tuple_spread_guard': 'tuple with differently represented elements',
    'detached_method_reattach': 'refuses a method read as a value',
    'number_extra_arguments_guard': 'error TS2345',
}

def capture(name, stage, command, environment=None):
    env = os.environ.copy()
    if environment:
        env.update(environment)
    completed = subprocess.run(command, cwd=root, env=env, capture_output=True, timeout=120)
    stem = results / (name + '.' + stage)
    stem.with_suffix(stem.suffix + '.stdout').write_bytes(completed.stdout)
    stem.with_suffix(stem.suffix + '.stderr').write_bytes(completed.stderr)
    record = {'command': command, 'environment': environment or {}, 'exit': completed.returncode,
              'stdout': completed.stdout.decode('utf-8', errors='backslashreplace'),
              'stderr': completed.stderr.decode('utf-8', errors='backslashreplace'),
              'stdout_hex': completed.stdout.hex(), 'stderr_hex': completed.stderr.hex()}
    stem.with_suffix(stem.suffix + '.json').write_text(json.dumps(record, indent=2) + '\n')
    return record

for source in sorted(unit.glob('*.a')):
    name = source.stem
    if args.names and name not in args.names:
        continue
    path = str(source.relative_to(root))
    runs = {'node': capture(name, 'node', ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', path])}
    for mode, flags in [('sanitized', ['--sanitize']), ('release', [])]:
        binary = str(work / (name + '.' + mode))
        build = capture(name, mode + '-build', [args.compiler, 'build', path, '-o', binary] + flags)
        runs[mode] = {'compile': build}
        if build['exit'] == 0:
            runs[mode]['run'] = capture(name, mode, [binary], {'ASAN_OPTIONS': 'detect_leaks=0', 'UBSAN_OPTIONS': 'halt_on_error=1'})
            if mode == 'sanitized' and runs[mode]['run']['exit'] == 0:
                runs['leaks'] = capture(name, 'leaks', [binary], {'ASAN_OPTIONS': 'detect_leaks=1:exitcode=23', 'LSAN_OPTIONS': 'exitcode=23'})
    build = capture(name, 'js-build', [args.compiler, 'js', path])
    runs['javascript'] = {'compile': build}
    if build['exit'] == 0:
        js = work / (name + '.mjs')
        js.write_text(build['stdout'])
        build['stdout'] = ''
        build['stdout_hex'] = ''
        build['generated_artifact'] = str(js)
        runs['javascript']['run'] = capture(name, 'javascript', ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(js)])
    observations = [runs['node']] + [runs[x].get('run') for x in ['sanitized', 'release', 'javascript']]
    def output(r):
        return (r['exit'], r['stdout_hex'], r['stderr_hex'])
    runs['agree'] = all(r is not None for r in observations) and len({output(r) for r in observations}) == 1
    runs['clean'] = runs['agree'] and runs['node']['exit'] == 0 and runs.get('leaks', {}).get('exit') == 0
    if name in expected_refusals:
        runs['check_passed'] = runs['node']['exit'] == 0 and all(
            runs[mode]['compile']['exit'] == 1 and expected_refusals[name] in runs[mode]['compile']['stderr']
            for mode in ['sanitized', 'release', 'javascript'])
    else:
        runs['check_passed'] = runs['agree'] and (runs['clean'] or name == 'uncaught_nested_throw' and runs['node']['exit'] == 70)
    records[name] = runs
    print(f"{name}: agree={runs['agree']} clean={runs['clean']}", flush=True)
(results / 'observations.json').write_text(json.dumps(records, indent=2) + '\n')

sys.exit(0 if all(r['check_passed'] for r in records.values()) else 1)
