"""Audit the inventory against the census and enforce the fixture status schema."""
import argparse
import collections
import hashlib
import json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('source', type=Path)
parser.add_argument('census', type=Path)
parser.add_argument('--mutants', action='store_true')
args = parser.parse_args()
bucket = Path(__file__).resolve().parent
inventory = json.loads((bucket / 'sites.json').read_text())
assert inventory['typescript'] == '6.0.3'

def recount(value):
    assert dict(collections.Counter(site['reason'] for site in value['sites'])) == value['counts'], 'source site recount differs'
    assert value['counts']['a spread after the first field'] == 17
    assert value['counts']['a definite assignment assertion !'] == 12
    expected = json.loads((args.census / 'data/shape_additions.json').read_text())
    key = lambda site: (site['file'], site['line'], site['column'], site['object'], site['property'], site['initial_kind'], site['declared_property'])
    assert sorted(map(key, value['post_creation_writes'])) == sorted(map(key, expected)), 'post-creation ledger differs'
    assert len(expected) == 16

recount(inventory)
files = json.loads((args.census / 'data/files.json').read_text())
checked = 0
for row in files:
    if row.get('generated'):
        continue
    assert hashlib.sha256((args.source / row['file']).read_bytes()).hexdigest() == row['sha256'], row['file']
    checked += 1
print(f'Pinned source hashes: {checked} original sources checked')
status = json.loads((bucket / 'status.json').read_text())
fixtures = {file.name for file in bucket.glob('*.a')}
assert len(status) == len(fixtures) == 25
assert {row['file'] for row in status} == fixtures
for row in status:
    assert set(row) == {'file', 'tsc', 'reason', 'node', 'stage0'}
    assert set(row['node']) == {'stdout', 'stderr', 'exit'}
    assert set(row['stage0']) == {'outcome', 'what'}
    assert row['stage0']['outcome'] in {'NotYet', 'Refused', 'Checker', 'Compiles'}
    assert row['node']['exit'] == 0 and row['node']['stderr'] == ''
    text = (bucket / row['file']).read_text()
    for location in row['tsc']:
        assert '// From TypeScript 6.0.3, ' + location in text
    assert '// Census reason: ' + row['reason'] in text
    assert (row['stage0']['what'] == '') == (row['stage0']['outcome'] == 'Compiles')
native = json.loads((bucket / 'native.json').read_text())
assert {row['file'] for row in native} == {row['file'] for row in status if row['stage0']['outcome'] == 'Compiles'}
for row in native:
    oracle = next(item['node'] for item in status if item['file'] == row['file'])
    assert row['matches_node'] and row['native'] == oracle, 'native differs from Node'
print('Status schema, attribution, native bytes: pass')
if args.mutants:
    changed = json.loads(json.dumps(inventory))
    index = next(i for i, row in enumerate(changed['sites']) if row['reason'] == 'a spread after the first field')
    del changed['sites'][index]
    try:
        recount(changed)
    except AssertionError as error:
        print('Dropped late-spread inventory site caught: ' + str(error))
    else:
        raise AssertionError('inventory mutant survived')
    mutant = json.loads((bucket / 'mutant.json').read_text())
    recorded = next(row['stage0'] for row in status if row['file'] == mutant['file'])
    assert mutant['before'] == recorded
    assert mutant['after'] != recorded and mutant['after']['outcome'] == 'Compiles'
    assert mutant['native'] == mutant['node']
    print('Plain-spread rewrite caught by recorded stage-0 outcome: Refused != Compiles')
