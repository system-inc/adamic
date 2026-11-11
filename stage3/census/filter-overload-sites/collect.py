#!/usr/bin/env python3
"""Select complete durable shard attempts, verify test identities, and sum site counters."""
import argparse
import collections
import hashlib
import json
from pathlib import Path
import sys

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('attempts', type=Path, help='JSON list of label/results/counters objects, oldest first')
parser.add_argument('sites', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('--require-complete', action='store_true')
args = parser.parse_args()
repository = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(repository / 'stage3/lane'))
from shard_plan import load_plan, names_digest
plan = load_plan()
units = {unit['name']: unit for unit in plan['units']}
source = json.loads(args.sites.read_text())
source = source['sites'] if isinstance(source, dict) else source
keys = ('calls', 'completed', 'identity_equal', 'returned_argument_array', 'undefined_argument')
totals = {row['where']: dict(id=row['id'], where=row['where'], **{key: 0 for key in keys}) for row in source}
attempts = []
for attempt in json.loads(args.attempts.read_text()):
    root = Path(attempt['results'])
    ledger_file = root / 'ledger.jsonl'
    ledger = [json.loads(line) for line in ledger_file.read_text().splitlines()] if ledger_file.exists() else []
    assert len({row['name'] for row in ledger}) == len(ledger), 'duplicate unit in attempt ledger'
    snapshots = collections.defaultdict(list)
    for file in sorted(Path(attempt['counters']).glob('*.json')):
        receipt = json.loads(file.read_text())
        unit = receipt.get('unit')
        if unit not in units:
            assert not receipt['sites'], 'unattributed executed filter calls'
            continue
        snapshots[unit].append(receipt)
    attempts.append(dict(**attempt, ledger={row['name']: row for row in ledger}, snapshots=snapshots))
selected = []
missing = []
processes = []
reports = []
for name, unit in units.items():
    accepted = None
    reasons = []
    for attempt in reversed(attempts):
        ledger = attempt['ledger'].get(name)
        if ledger is None:
            continue
        receipts = attempt['snapshots'].get(name, [])
        if ledger['exit'] or ledger['status'] != 'pass':
            reasons.append(dict(attempt=attempt['label'], reason='unit did not pass', errors=ledger.get('errors')))
            continue
        if not receipts or any(row.get('format') != 2 or not row['final'] for row in receipts):
            reasons.append(dict(attempt=attempt['label'], reason='missing, legacy, or incomplete counter receipt'))
            continue
        root = Path(attempt['results']) / name
        report = json.loads((root / 'report.json').read_text())
        observations = json.loads((root / 'observations.json').read_text())
        assert names_digest(observations) == unit['names_digest'], 'test identity mismatch: ' + name
        counts = {key: sum(row['status'] == status for row in observations) for key, status in [('pass', 'pass'), ('fail', 'fail'), ('skip', 'pending')]}
        assert counts == report['counts'] == unit['expect'] == ledger['counts'], 'test count mismatch: ' + name
        assert report['status'] == 'pass' and not report['errors']
        accepted = attempt, ledger, report, receipts, counts
        break
    if accepted is None:
        missing.append(dict(name=name, reasons=reasons))
        continue
    attempt, ledger, report, receipts, counts = accepted
    for receipt in receipts:
        processes.append(dict(attempt=attempt['label'], **receipt))
        for row in receipt['sites']:
            assert row['id'] in totals, 'unexpected instrumented site: ' + row['id']
            assert all(isinstance(row[key], int) and row[key] >= 0 for key in keys)
            assert row['returned_argument_array'] <= row['identity_equal'] <= row['completed'] <= row['calls']
            assert row['undefined_argument'] <= row['calls']
            for key in keys:
                totals[row['id']][key] += row[key]
    selected.append(dict(name=name, attempt=attempt['label'], seconds=ledger['seconds'], counts=counts,
                         pids=[row['pid'] for row in receipts], inputs_hash=ledger['inputs_hash'],
                         names_digest=unit['names_digest']))
    reports.append(dict(attempt=attempt['label'], **report))
result = dict(complete=len(selected) == len(units), expected_units=len(units), selected_units=len(selected),
              test_counts={key: sum(row['counts'][key] for row in selected) for key in ('pass', 'fail', 'skip')},
              observer_sha256=hashlib.sha256((Path(__file__).parent / 'observe.cjs').read_bytes()).hexdigest(),
              manifest_sha256=hashlib.sha256((repository / 'stage3/lane/shards.json').read_bytes()).hexdigest(),
              totals={key: sum(row[key] for row in totals.values()) for key in keys},
              sites=list(totals.values()), selected=selected, missing=missing,
              selected_processes=processes, selected_reports=reports)
args.output.write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({key: result[key] for key in ('complete', 'expected_units', 'selected_units', 'test_counts', 'totals')}))
raise SystemExit(int(args.require_complete and not result['complete']))
