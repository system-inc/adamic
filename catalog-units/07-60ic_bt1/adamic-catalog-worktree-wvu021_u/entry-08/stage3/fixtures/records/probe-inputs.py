"""Run committed real inputs with unmodified stock TypeScript 6.0.3 on Node."""
import argparse
import json
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('--observe', action='store_true')
args = parser.parse_args()
package = Path(os.environ['STOCK_TYPESCRIPT'])
assert json.loads((package / 'package.json').read_text())['version'] == '6.0.3'
cases = json.loads((ROOT / 'input-fixtures/manifest.json').read_text())
results = []
with tempfile.TemporaryDirectory(prefix='records-inputs-') as temporary:
    for case in cases:
        directory = Path(temporary) / case['name']
        shutil.copytree(ROOT / 'input-fixtures' / case['name'], directory)
        for source in directory.rglob('*.a'):
            source.rename(source.with_suffix('.ts'))
        command = ['node', str(package / 'lib/tsc.js'), *case['args']]
        run = subprocess.run(command, cwd=directory, capture_output=True, text=True, timeout=30)
        expected = case['expected_diagnostic']
        correct = (run.returncode == 0 and run.stdout == '' and run.stderr == '') if expected == 0 else (run.returncode == case['expected_exit'] and f'error TS{expected}:' in run.stdout and run.stderr == '')
        behavior = 'correct' if correct else ('crash' if run.stderr else ('silently ignoring it' if run.returncode == 0 else 'wrong diagnostic'))
        result = {**case, 'behavior': behavior, 'node': {'stdout': run.stdout, 'stderr': run.stderr, 'exit': run.returncode}}
        results.append(result)
        print(case['name'], behavior, 'exit=' + str(run.returncode), repr(run.stdout))
    # Real-input mutant: ordinary own constructor path resolves, dropped __proto__ does not.
    own = next(r for r in results if r['form'] == 'paths-own' and r['key'] == '__proto__')
    mutant = next(r for r in results if r['form'] == 'paths-own' and r['key'] == 'constructor')
    assert own['node'] != mutant['node']
    print('CAUGHT by stock CLI output: replace lost own __proto__ path with ordinary own constructor path')
status = ROOT / 'input-fixtures/observations.json'
if args.observe:
    status.write_text(json.dumps({'typescript': '6.0.3', 'cases': results}, indent=2) + '\n')
else:
    assert results == json.loads(status.read_text())['cases']
print('PASS:', len(results), 'exact stock CLI observations')
