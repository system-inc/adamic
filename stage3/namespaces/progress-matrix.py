"""Append a fresh Node/native/checked-JavaScript row for all 12 upstream slices."""
from pathlib import Path
import argparse
import json
import os
import re
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--label', required=True)
parser.add_argument('--scratch', type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parents[2]
scratch = args.scratch.resolve()
scratch.mkdir(parents=True, exist_ok=True)

def run(command, name, env=None):
    process = subprocess.run(command, cwd=root, capture_output=True, env=env, timeout=300)
    result = {'stdout': process.stdout.decode(), 'stderr': process.stderr.decode(), 'exit': process.returncode}
    (scratch / (name + '.json')).write_text(json.dumps(result, indent=2) + '\n')
    return result

compiler = scratch / 'adamic'
build = run(['go', 'build', '-o', str(compiler), './cmd/adamic'], 'compiler')
assert build['exit'] == 0, build
row = {'label': args.label, 'commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip(), 'fixtures': []}
fixtures = sorted((root / 'stage3/fixtures/namespaces').glob('[0-9][0-9]_*.a'))
assert len(fixtures) == 12
for fixture in fixtures:
    node = run(['node', '--disable-warning=ExperimentalWarning', str(root / 'oracle/node.mjs'), str(fixture)], fixture.stem + '.node')
    assert node['exit'] == 0 and node['stderr'] == '', node
    binary = scratch / fixture.stem
    build = run([str(compiler), 'build', str(fixture), '-o', str(binary), '--sanitize'], fixture.stem + '.build')
    if build['exit'] == 0:
        kind = 'Compiles'
    elif "stage 0 can't lower" in build['stderr']:
        kind = 'NotYet'
    elif 'Adamic 0.1 refuses' in build['stderr']:
        kind = 'Refused'
    elif re.search(r'error TS[0-9]+:', build['stderr']):
        kind = 'Checker'
    else:
        raise RuntimeError(build)
    result = {'file': fixture.name, 'node': node, 'outcome': kind, 'build': build}
    if kind == 'Compiles':
        result['native'] = run([str(binary)], fixture.stem + '.native', dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1'))
        javascript = run([str(compiler), 'js', str(fixture)], fixture.stem + '.js-build')
        assert javascript['exit'] == 0, javascript
        path = scratch / (fixture.stem + '.mjs')
        path.write_text(javascript['stdout'])
        result['checkedJavaScript'] = run(['node', '--disable-warning=ExperimentalWarning', str(root / 'oracle/node.mjs'), str(path)], fixture.stem + '.checked-js')
        for backend in ['native', 'checkedJavaScript']:
            if result[backend] != node:
                raise RuntimeError((fixture.name, backend, 'Node disagreement', node, result[backend]))
    row['fixtures'].append(result)
row['counts'] = {kind: sum(result['outcome'] == kind for result in row['fixtures']) for kind in ['Compiles', 'NotYet', 'Refused', 'Checker']}
output = root / 'stage3/namespaces/progress-matrix.json'
rows = json.loads(output.read_text()) if output.exists() else []
rows.append(row)
output.write_text(json.dumps(rows, indent=2) + '\n')
print(args.label, row['commit'][:7], row['counts'])
