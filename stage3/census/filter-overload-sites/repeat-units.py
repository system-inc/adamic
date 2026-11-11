#!/usr/bin/env python3
"""Repeat named units through unchanged manifest commands; retain each child's exit."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('units', type=Path)
parser.add_argument('results', type=Path)
args = parser.parse_args()
repository = Path(__file__).resolve().parents[3]
plan = json.loads((repository / 'stage3/lane/shards.json').read_text())
wanted = json.loads(args.units.read_text())
assert len(wanted) == len(set(wanted))
units = {unit['name']: unit for unit in plan['units']}
assert set(wanted) <= units.keys()
root = args.results.resolve()
root.mkdir(parents=True, exist_ok=False)
rows = []
for name in wanted:
    unit = units[name]
    env = dict(os.environ, STAGE3_RESULTS=str(root), STAGE3_CACHE=str(root / 'caches' / name),
               NODE_DISABLE_COMPILE_CACHE='1', PYTHONDONTWRITEBYTECODE='1')
    started = time.monotonic()
    with (root / (name + '.log')).open('w') as log:
        child = subprocess.run(unit['command'], cwd=repository, env=env, stdout=log, stderr=subprocess.STDOUT)
    report_file = root / name / 'report.json'
    report = json.loads(report_file.read_text()) if report_file.exists() else {}
    row = dict(name=name, exit=child.returncode, seconds=time.monotonic()-started,
               status=report.get('status', 'missing'), counts=report.get('counts'),
               errors=report.get('errors'), inputs_hash=report.get('inputs_hash'), artifact=report.get('artifact'))
    rows.append(row)
    with (root / 'ledger.jsonl').open('a') as stream:
        stream.write(json.dumps(row) + '\n')
    print(json.dumps(row), flush=True)
(root / 'ledger.json').write_text(json.dumps(dict(units=rows), indent=2) + '\n')
raise SystemExit(int(any(row['exit'] for row in rows)))
