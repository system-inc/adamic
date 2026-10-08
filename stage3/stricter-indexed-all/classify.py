#!/usr/bin/env python3
"""Classify measured whole-program diagnostics by exact ledger identity."""
import argparse
import csv
import json
import re
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--tree', type=Path, required=True)
parser.add_argument('--strict', type=Path, required=True)
parser.add_argument('--production', type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parent
evidence = root / 'evidence'
rows = list(csv.DictReader((evidence / 'ledger-rows.csv').open()))
indexed = [row for row in rows if row['option'] == 'noUncheckedIndexedAccess']
if len(indexed) != 99:
    raise SystemExit(f'expected 99 indexed rows, got {len(indexed)}')

def key(row):
    return row['file'], int(row['line']), int(row['column']), row['code']

def diagnostics(profile):
    parsed = {}
    for text in profile['diagnostics']:
        match = re.fullmatch(r'(.+):(\d+):(\d+): error (TS\d+): ([\s\S]*)', text)
        if not match:
            raise SystemExit('unrecognized diagnostic: ' + text)
        file, line, column, code, message = match.groups()
        relative = str(Path(file).relative_to(args.tree))
        identity = relative, int(line), int(column), code
        if identity in parsed:
            raise SystemExit('duplicate diagnostic identity: ' + str(identity))
        parsed[identity] = message
    return parsed

strict = json.loads(args.strict.read_text())
production = json.loads(args.production.read_text())
for profile in [strict, production]:
    if len(profile['roots']) != 79:
        raise SystemExit('whole-program root coverage changed')
strict_errors = diagnostics(strict)
production_errors = diagnostics(production)
known = {key(row): row for row in rows}
extras = set(strict_errors) - known.keys()
if extras:
    raise SystemExit('unattributed strict errors: ' + str(extras))
# This report describes the two observed checker-blocked runs. Refuse to
# silently classify a later successful/lower-stage run with this old model.
if any(profile['stage'] != 'checker' or profile.get('binary') for profile in [strict, production]):
    raise SystemExit('new whole-program stage requires a new native-outcome classification')
if set(key(row) for row in indexed) - strict_errors.keys():
    raise SystemExit('strict profile is missing an expected indexed diagnostic')
if set(key(row) for row in indexed) & production_errors.keys():
    raise SystemExit('production conversion retained an indexed checker diagnostic')
if set(production_errors) - known.keys():
    raise SystemExit('unattributed production errors')
minimal = {}
for path in sorted((root.parent / 'stricter-indexed-a').glob('D*.json')):
    fixture = json.loads(path.read_text())
    for identity in fixture['ids']:
        minimal[identity] = {'slice': 'a', 'state': 'refused' if fixture.get('block') else 'proven', 'reason': fixture.get('block', '')}
for row in json.loads((root.parent / 'stricter-indexed-b/sites.json').read_text()):
    minimal[row['id']] = {'slice': 'b', 'state': 'proven', 'reason': ''}
for fixture in json.loads((root.parent / 'stricter-indexed-c/sites.json').read_text()):
    minimal[fixture['id']] = {'slice': 'c', 'state': 'refused' if fixture.get('blocked') else 'proven', 'reason': fixture.get('blocked', '')}
for fixture in json.loads((root.parent / 'stricter-indexed-d/sites.json').read_text()):
    minimal[fixture['id']] = {'slice': 'd', 'state': 'refused' if fixture['blocked'] else 'proven', 'reason': fixture.get('refusal', 'Adamic 0.1 refuses an index signature; use a Map') if fixture['blocked'] else ''}
if set(minimal) != {row['id'] for row in indexed}:
    raise SystemExit('slice row coverage differs from the exact 99-row ledger')
results = []
for row in sorted(indexed, key=lambda row: int(row['id'][1:])):
    identity = key(row)
    actual = dict(row)
    actual['strict_state'] = 'checker error' if identity in strict_errors else 'not directly diagnosed'
    actual['strict_reason'] = strict_errors.get(identity, '')
    actual['production_state'] = 'checker error' if identity in production_errors else 'deferred, lowering not reached'
    actual['production_reason'] = production_errors.get(identity, f"whole-program {production['stage']} rejected before this read: {len(production['diagnostics'])} other diagnostics")
    actual['minimal_witness'] = minimal[row['id']]
    results.append(actual)
(evidence / 'indexed-status.json').write_text(json.dumps(results, indent=2) + '\n')
fields = ['id', 'file', 'line', 'column', 'code', 'owner', 'source_expression', 'strict_state', 'strict_reason', 'production_state', 'production_reason']
with (evidence / 'indexed-status.csv').open('w', newline='') as output:
    writer = csv.DictWriter(output, fieldnames=fields, extrasaction='ignore', lineterminator='\n')
    writer.writeheader()
    writer.writerows(results)
summary = {
    'roots': 79,
    'literal_strict': {'stage': strict['stage'], 'diagnostics': len(strict_errors), 'indexed_errors': sum(key(row) in strict_errors for row in indexed)},
    'production_conversion': {'stage': production['stage'], 'diagnostics': len(production_errors), 'indexed_errors': sum(key(row) in production_errors for row in indexed), 'native_indexed_outcomes_unmeasured': 99},
    'minimal_manifest_dispositions': {'proven': sum(value['state'] == 'proven' for value in minimal.values()), 'refused': sum(value['state'] == 'refused' for value in minimal.values())},
    'native_artifact': None,
}
(evidence / 'whole-summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary))
