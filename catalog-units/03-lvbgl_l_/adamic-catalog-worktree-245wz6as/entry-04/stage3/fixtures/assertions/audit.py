"""Audit census coverage, ledger totals, extraction fidelity, and status schema."""
import argparse
import collections
import copy
import json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('census', type=Path)
parser.add_argument('source', type=Path)
parser.add_argument('--mutants', action='store_true')
args = parser.parse_args()
bucket = Path(__file__).resolve().parent
sites = json.loads((args.census / 'sites.json').read_text())
ledger = json.loads((bucket / 'ledger.json').read_text())
summary = json.loads((bucket / 'ledger-summary.json').read_text())
nonnull = json.loads((bucket / 'nonnull-ledger.json').read_text())
manifest = json.loads((bucket / 'manifest.json').read_text())
status = json.loads((bucket / 'status.json').read_text())

def key(row):
    return row['file'], row['start'], row['end']

def check_ledger(rows, totals):
    expected = [s for s in sites if s['reason'] == 'type assertion sites' and '.generated.' not in s['file']]
    assert sorted(map(key, rows)) == sorted(map(key, expected)), 'assertion location coverage'
    assert len(set(map(key, rows))) == len(rows), 'duplicate locations'
    assert len(rows) == totals['assertions'] == 4101, 'assertion total'
    assert dict(collections.Counter(s['category'] for s in rows)) == totals['counts'], 'category totals'

check_ledger(ledger, summary)
expected_bangs = [s for s in sites if s['reason'] == 'the non-null assertion !' and '.generated.' not in s['file']]
assert sorted(map(key, nonnull)) == sorted(map(key, expected_bangs))
assert len(nonnull) == 1123
assert dict(collections.Counter(s['form'] for s in nonnull)) == summary['nonnull_counts']
assert len(manifest) == len(status) == 20
assert set(r['file'] for r in manifest) == set(r['file'] for r in status)
for row in status:
    assert set(row) == {'file', 'tsc', 'reason', 'node', 'stage0'}
    assert set(row['node']) == {'stdout', 'stderr', 'exit'}
    assert set(row['stage0']) == {'outcome', 'what'}
    assert row['stage0']['outcome'] in {'NotYet', 'Refused', 'Checker', 'Compiles'}
    assert row['node']['exit'] == 0 and row['node']['stderr'] == ''
    if row['stage0']['outcome'] == 'Compiles':
        assert row['stage0']['what'] == ''
        recorded = json.loads((bucket / 'logs' / (row['file'] + '.json')).read_text())
        assert recorded['native'] == recorded['node'] == row['node'], 'native differs from Node'
for row in manifest:
    name, start = row['tsc'][0].rsplit(':', 1)
    lines = (args.source / name).read_text().splitlines()
    span = lines[int(start)-1:row['source_end_line']]
    original = '\n'.join(span).strip()
    # Original indentation is present except before the first extracted token.
    if row['function'] is not None:
        first = span[0].lstrip()
        original = '\n'.join([first] + span[1:]).strip()
    assert original in (bucket / row['file']).read_text(), 'changed upstream span: ' + row['file']
print('PASS: 4101 assertions, 1123 non-null sites, 20 source spans and status rows')
if args.mutants:
    for name, rows, totals, catch in [
        ('omitted assertion', ledger[:-1], summary, 'assertion location coverage'),
        ('inflated category count', ledger, copy.deepcopy(summary), 'category totals'),
    ]:
        if name == 'inflated category count':
            totals['counts']['upcast'] += 1
        try:
            check_ledger(rows, totals)
        except AssertionError as error:
            assert str(error) == catch, (name, str(error))
            print('CAUGHT:', name, 'by', catch)
        else:
            raise AssertionError('surviving mutant: ' + name)
