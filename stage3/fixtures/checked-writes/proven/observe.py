#!/usr/bin/env python3
"""Measure this bucket only. Logs and binaries stay outside the repository."""
import json
import pathlib
import subprocess
import sys

bucket = pathlib.Path(__file__).resolve().parent
repository = bucket.parents[3]
logs = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else '/tmp/step10-observations')
logs.mkdir(parents=True, exist_ok=True)
ledger = json.loads((bucket / 'sites.json').read_text())
rows = []
for fixture in ledger['fixtures']:
    file = bucket / fixture['file']
    name = file.stem
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(file)], cwd=repository, capture_output=True)
    (logs / (name + '.node.stdout')).write_bytes(node.stdout)
    (logs / (name + '.node.stderr')).write_bytes(node.stderr)
    if node.returncode != 0:
        raise RuntimeError(f'{name}: Node failed: {node.stderr.decode()}')
    binary = logs / (name + '.native')
    build = subprocess.run(['go', 'run', './cmd/adamic', 'build', str(file), '-o', str(binary)], cwd=repository, capture_output=True)
    (logs / (name + '.build.log')).write_bytes(build.stdout + build.stderr)
    what = (build.stdout + build.stderr).decode()
    if build.returncode == 0:
        native = subprocess.run([str(binary)], capture_output=True)
        (logs / (name + '.native.stdout')).write_bytes(native.stdout)
        (logs / (name + '.native.stderr')).write_bytes(native.stderr)
        if (native.stdout, native.stderr, native.returncode) != (node.stdout, node.stderr, node.returncode):
            raise RuntimeError(f'SILENT MISCOMPILE: {name}; inspect {logs}')
        stage0 = {'outcome': 'Compiles', 'what': ''}
    else:
        outcome = 'NotYet' if 'not yet' in what.lower() or 'notyet' in what.lower() or "can't lower" in what.lower() else 'Refused' if 'refused' in what.lower() or 'refuses' in what.lower() else 'Checker'
        stage0 = {'outcome': outcome, 'what': what}
    row = next(row for row in ledger['records'] if row['id'] == fixture['representative_id'])
    rows.append({'file': fixture['file'], 'tsc': [span['file'] + ':' + str(span['start']) for span in fixture['spans']], 'reason': row['reason'], 'node': {'stdout': node.stdout.decode(), 'stderr': node.stderr.decode(), 'exit': node.returncode}, 'stage0': stage0})
(bucket / 'status.json').write_text(json.dumps(rows, indent=2) + '\n')
print(f'{len(rows)} Node observations; stage0 ' + ', '.join(f'{outcome}={sum(row["stage0"]["outcome"] == outcome for row in rows)}' for outcome in ['Compiles', 'NotYet', 'Refused', 'Checker']))
