"""Run the external source oracle and current stage 0; keep all process output."""
import json
import subprocess
from pathlib import Path

bucket = Path(__file__).resolve().parent
repository = bucket.parents[2]
rows = json.loads((bucket / 'manifest.json').read_text())
native = []
for row in rows:
    fixture = str((bucket / row['file']).relative_to(repository))
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', fixture], cwd=repository, capture_output=True)
    row['node'] = {'stdout': node.stdout.decode(), 'stderr': node.stderr.decode(), 'exit': node.returncode}
    binary = '/tmp/objects-' + row['file'] + '.bin'
    stage0 = subprocess.run(['go', 'run', './cmd/adamic', 'build', fixture, '-o', binary], cwd=repository, capture_output=True)
    diagnostic = stage0.stderr.decode().removesuffix('exit status 1\n')
    if stage0.returncode == 0:
        outcome = 'Compiles'
        result = subprocess.run([binary], cwd=repository, capture_output=True)
        observed = {'stdout': result.stdout.decode(), 'stderr': result.stderr.decode(), 'exit': result.returncode}
        matches = node.stdout == result.stdout and node.stderr == result.stderr and node.returncode == result.returncode
        native.append({'file': row['file'], 'native': observed, 'matches_node': matches})
    elif 'refuses' in diagnostic:
        outcome = 'Refused'
    elif "can't lower" in diagnostic:
        outcome = 'NotYet'
    elif 'error TS' in diagnostic:
        outcome = 'Checker'
    else:
        raise RuntimeError('Unclassified compiler failure: ' + diagnostic)
    row['stage0'] = {'outcome': outcome, 'what': diagnostic}
    print(json.dumps({'file':row['file'], 'node_exit':node.returncode, 'stage0':row['stage0']}), flush=True)
    if node.returncode:
        raise RuntimeError('Fixture fails on Node: ' + fixture)
(bucket / 'status.json').write_text(json.dumps(rows, indent=2) + '\n')
(bucket / 'native.json').write_text(json.dumps(native, indent=2) + '\n')
if any(not row['matches_node'] for row in native):
    raise RuntimeError('SILENT MISCOMPILE: see native.json')
print(f'{len(rows)} fixtures, {len(native)} native comparisons, all match', flush=True)
