#!/usr/bin/env python3
"""Run source Node, sanitized native, release native and emitted JavaScript."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
unit = Path(__file__).resolve().parent
compiler = os.environ.get('ADAMIC_COVERAGE_COMPILER', '/tmp/trainone-adamic')
evidence = unit / 'evidence'
evidence.mkdir(exist_ok=True)
summary = {}
expected_stops = {'entries-absent', 'entries-delete', 'entries-presence', 'optional-map-nullish', 'optional-nested-nullish', 'optional-record', 'substr-omitted-start'}

def run(name, argv, env=None):
    with (evidence / (name + '.stdout')).open('wb') as out, (evidence / (name + '.stderr')).open('wb') as err:
        process = subprocess.run(argv, cwd=root, stdout=out, stderr=err, env=env, timeout=120)
    result = {'command': argv, 'exit': process.returncode,
              'stdout': (evidence / (name + '.stdout')).read_bytes().decode('utf-8', 'backslashreplace'),
              'stderr': (evidence / (name + '.stderr')).read_bytes().decode('utf-8', 'backslashreplace')}
    (evidence / (name + '.json')).write_text(json.dumps(result, indent=2) + '\n')
    return result

with tempfile.TemporaryDirectory(prefix='trainone-', dir='/tmp/adamic-gate') as scratch:
    for path in sorted(unit.glob('*.a')):
        name = path.stem
        results = {'node': run(name + '-node', ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(path)])}
        for mode in ['sanitize', 'O2', 'javascript']:
            output = str(Path(scratch) / (name + '-' + mode + ('.mjs' if mode == 'javascript' else '')))
            argv = [compiler, 'js', str(path)] if mode == 'javascript' else [compiler, 'build', str(path), '-o', output] + (['--sanitize'] if mode == 'sanitize' else [])
            built = run(name + '-' + mode + '-compile', argv)
            if built['exit'] != 0:
                results[mode] = dict(built, phase='compile')
                continue
            if mode == 'javascript':
                Path(output).write_bytes((evidence / (name + '-' + mode + '-compile.stdout')).read_bytes())
                argv = ['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', output]
            else:
                argv = [output]
            env = dict(os.environ, ASAN_OPTIONS='detect_leaks=0', UBSAN_OPTIONS='halt_on_error=1:print_stacktrace=1')
            results[mode] = run(name + '-' + mode, argv, env)
            if mode == 'sanitize' and results[mode]['exit'] == 0:
                env['ASAN_OPTIONS'] = 'detect_leaks=1'
                results['leaks'] = run(name + '-leaks', argv, env)
        triples = [(results[m]['stdout'], results[m]['stderr'], results[m]['exit']) for m in ['node', 'sanitize', 'O2', 'javascript']]
        results['agree'] = all(t == triples[0] for t in triples) and all(results[m].get('phase') != 'compile' for m in ['sanitize', 'O2', 'javascript'])
        summary[name] = results
        print(name, 'agree=' + str(results['agree']), {m: results[m]['exit'] for m in ['node', 'sanitize', 'O2', 'javascript']}, flush=True)
(evidence / 'results.json').write_text(json.dumps(summary, indent=2) + '\n')

for name, results in summary.items():
    if name in expected_stops:
        assert results['node']['exit'] == 0 and results['node']['stderr'] == '', name
        assert all(results[m].get('phase') == 'compile' and results[m]['exit'] == 1 for m in ['sanitize', 'O2', 'javascript']), name
    else:
        assert results['agree'] and results['node']['exit'] == 0 and results['node']['stderr'] == '', name
        assert results['leaks']['exit'] == 0 and results['leaks']['stderr'] == '', name
print('All supported comparisons and explicit compiler stops matched expectations.')
