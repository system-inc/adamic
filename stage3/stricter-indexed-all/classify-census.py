#!/usr/bin/env python3
"""Retain production ledger states without calling an unvisited read refused."""
import argparse
from collections import Counter
import csv
import json
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--result', type=Path, required=True)
parser.add_argument('--tree', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
root = Path(__file__).resolve().parent
manifest = list(csv.DictReader((root / 'evidence/ledger-rows.csv').open()))
result = json.loads(args.result.read_text())
census = result['census']
assert len(manifest) == len(census['rows']) == 173
assert set(census['rows']) == {row['id'] for row in manifest}
assert census['ordinary_errors'] == []
observed_counts = Counter()
for row in census['rows'].values():
    observed_counts[row['state']] += 1
    observed_counts[row['state'] + ':' + row['kind']] += 1
assert dict(observed_counts) == census['counts']
assert len(result['roots']) == 79
assert result['stage'] == 'lower' and not result['diagnostics']
assert result['indexed_audit_rows'] == 99
assert result['requires_indexed_presence']
identity = lambda site: (site['file'], site['line'], site['column'], site['code'], tuple(site['options']))
assert {identity(site) for site in result['option_sites']} == {identity(row['site']) for row in census['rows'].values()}
assert len(result['option_sites']) == 173
expected_exclusions = {row['id'] for row in manifest if row['option'] in ('exactOptionalPropertyTypes', 'useUnknownInCatchVariables')}
assert len(result['excluded']) == 72
assert {row['id'] for row in result['excluded']} == expected_exclusions
for row in result['excluded']:
    assert identity(row['site']) == identity(census['rows'][row['id']]['site'])
assert len(result['attempts']) == 16
# A future emitted artifact needs original source-to-guard attribution. Fail
# closed rather than treating a scheduled diagnostic as an emitted guard.
for attempt in result['attempts']:
    assert attempt['stage'] == 'lower' and attempt['error_kind'] in ('NotYet', 'Refused'), 'new lowering outcome requires source-to-guard attribution'
    assert attempt['checks'] == [] and not attempt.get('binary')
entry = next(attempt for attempt in result['attempts'] if attempt['entry'] == 'src/compiler/_namespaces/ts.ts')
rows = []
counts = {'scheduled': 0, 'emitted as a check': 0, 'refused': 0}
for original in manifest:
    decision = census['rows'][original['id']]
    site = decision['site']
    assert Path(site['file']) == args.tree.resolve() / original['file']
    assert (site['line'], site['column'], site['code']) == (int(original['line']), int(original['column']), int(original['code'].removeprefix('TS')))
    assert original['option'] in site['options'] or original['id'] in ('D112', 'D154') and original['option'] == 'other' and site['options'] == ['JSON.stringify']
    assert decision['state'] in ('scheduled-check', 'remaining-error')
    expected_kind = {'noUncheckedIndexedAccess': 'indexed-presence', 'exactOptionalPropertyTypes': 'optional-presence', 'useUnknownInCatchVariables': 'catch-error', 'other': 'json-stringify-defined'}[original['option']]
    assert decision['kind'] == expected_kind
    status = 'scheduled' if decision['state'] == 'scheduled-check' else 'refused'
    counts[status] += 1
    rows.append({**original, 'status': status, 'kind': decision['kind'], 'reason': decision.get('reason', ''), 'read_local_outcome': 'not reached', 'emitted_check': False, 'program_blocker': entry['error'], 'refusal_scope': 'checker diagnostic' if status == 'refused' else '', 'diagnostic_excluded_for_probe': status == 'refused'})
assert counts == {'scheduled': 168, 'emitted as a check': 0, 'refused': 5}
assert sum(row['option'] == 'noUncheckedIndexedAccess' and row['status'] == 'scheduled' for row in rows) == 99
assert {row['id'] for row in rows if row['status'] == 'refused'} == {'D108', 'D128', 'D132', 'D152', 'D153'}
args.output.mkdir(parents=True, exist_ok=True)
(args.output / 'census-row-outcomes.json').write_text(json.dumps(rows, indent=2) + '\n')
with (args.output / 'census-row-outcomes.csv').open('w') as output:
    writer = csv.DictWriter(output, fieldnames=list(rows[0]))
    writer.writeheader()
    writer.writerows(rows)
summary = {'rows': 173, 'states': counts, 'indexed': {'scheduled': 99, 'emitted as a check': 0, 'read-local refused': 0, 'read-local errors': 0, 'read-local not reached': 99}, 'ordinary_errors': 0, 'program_blocker': entry['error'], 'native_artifacts': 0, 'attempts': result['attempts']}
(args.output / 'census-summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary))
