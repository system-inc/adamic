#!/usr/bin/env python3
"""Check coverage, input hashes and observed check totals across the split."""
import csv
import hashlib
import json
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parents[2]
review = root / 'review/test-split-flow'
before = json.loads((review/'before.json').read_text())
after = json.loads((review/'after.json').read_text())
units = json.loads((review/'units.json').read_text())
old = {(r['package'],r['test']):r for r in before['tests']}
new = {(r['package'],r['test']):r for r in after['tests']}
assert len(old) == len(before['tests']) == 205
assert len(new) == len(after['tests']) == 1563
assert all(r['exit'] == 0 and r['elapsed'] < 30 for r in new.values())
assert sum(r['action'] == 'skip' for r in new.values()) == 2
families = {
 ('flow','TestEveryFunctionIsInSingleAssignment'): 'SingleAssignment',
 ('flow','TestEveryMutationIsInItsRange'): 'MutationRanges',
 ('flow','TestEveryPathNodeTakesIsInTheGraph'): 'GraphPaths',
 ('flow','TestLivenessHoldsOnEveryPath'): 'Liveness',
 ('fresh','TestEveryWriteIsRecordedAndKnown'): 'Writes',
}
expected = set(old) - set(families)
expected |= {(r['package'],r['test']) for r in units}
expected |= {('flow','TestFlowCorpusUnitsCoverEveryProgram'),('flow','TestFlowCorpusSetupIsShared'),('fresh','TestFreshCorpusUnitsCoverEveryProgram')}
assert set(new) == expected, (set(new)-expected, expected-set(new))
assert len(units) == 3385
for (package,test), family in families.items():
    pieces = [r for r in units if r['package'] == package and r['family'] == family]
    assert len(pieces) == (676 if package == 'flow' else 681)
    assert len({r['program'] for r in pieces}) == len(pieces)
    for r in pieces:
        assert before['inputs'][r['program']] == after['inputs'][r['program']] == r['sha256']
        assert hashlib.sha256((root/r['program']).read_bytes()).hexdigest() == r['sha256']
for path, digest in before['inputs'].items():
    if not path.endswith('_test.go'):
        assert after['inputs'][path] == digest, ('production or fixture changed', path)

for path, digest in after['inputs'].items():
    assert hashlib.sha256((root/path).read_bytes()).hexdigest() == digest, ('measured input changed', path)

def events(row):
    for line in Path(row['log']).read_text().splitlines():
        try:
            yield json.loads(line)
        except json.JSONDecodeError:
            pass

def invocation(row):
    terminal = [e for e in events(row) if 'Test' not in e and e['Action'] in ('pass','fail','skip')]
    assert len(terminal) == 1
    return terminal[0]['Elapsed']

assert all(invocation(r) < 30 for r in new.values()), 'an invocation exceeded the budget'

patterns = {
 'SingleAssignment': r'(\d+) functions: (\d+) phis, (\d+) values, (\d+) uses checked',
 'MutationRanges': r'(\d+) mutations checked, (\d+) on a range the pass left unset, (\d+) invalid ranges',
 'GraphPaths': r'(\d+) points walked, (\d+) events',
 'Liveness': r'(\d+) variable-and-point pairs checked',
 'Writes': r'(\d+) writes(?: in \d+ programs)?, (\d+) proven not to close a cycle',
}
def totals(rows, family):
    result = None
    matches = 0
    for row in rows:
        for event in events(row):
            match = re.search(patterns[family], event.get('Output',''))
            if match:
                values = [int(v) for v in match.groups()]
                if result is None:
                    result = [0]*len(values)
                result = [a+b for a,b in zip(result,values)]
                matches += 1
    assert result is not None, family
    return result, matches

observations = {}
for key, family in families.items():
    owned = {(r['package'],r['test']) for r in units if r['package'] == key[0] and r['family'] == family}
    previous, before_rows = totals([old[key]], family)
    current, after_rows = totals([new[k] for k in owned], family)
    assert previous == current, (family, previous, current)
    observations[family] = {'before': previous, 'after': current, 'before_log_rows': before_rows, 'after_log_rows': after_rows}

rows = []
for key, row in old.items():
    owned = {(r['package'],r['test']) for r in units if r['package'] == key[0] and r['family'] == families[key]} if key in families else {key}
    selected = [new[k] for k in owned]
    rows.append({'package':key[0], 'test':key[1], 'before_test_seconds':row['elapsed'],
                 'before_invocation_seconds':invocation(row), 'after_max_test_seconds':max(r['elapsed'] for r in selected),
                 'after_max_invocation_seconds':max(invocation(r) for r in selected), 'units':len(owned),
                 'before_status':row['action'], 'after_status': 'skip' if all(r['action']=='skip' for r in selected) else 'pass'})
for key in sorted(set(new)-set(old)-{(r['package'],r['test']) for r in units}):
    r = new[key]
    rows.append({'package':key[0], 'test':key[1], 'before_test_seconds':'', 'before_invocation_seconds':'',
                 'after_max_test_seconds':r['elapsed'], 'after_max_invocation_seconds':invocation(r), 'units':1,
                 'before_status':'new', 'after_status':r['action']})
with (review/'timings.csv').open('w') as f:
    writer = csv.DictWriter(f,fieldnames=list(rows[0]), lineterminator='\n')
    writer.writeheader()
    writer.writerows(rows)
summary = {'before_tests':len(old), 'after_tests':len(new), 'passed':sum(r['action']=='pass' for r in new.values()),
           'skipped':[f"{r['package']}/{r['test']}" for r in new.values() if r['action']=='skip'],
           'analysis_pieces':len(units), 'corpus_units':len({(r['package'],r['test']) for r in units}),
           'before_over_30_test':[r for r in rows if r['before_test_seconds']!='' and r['before_test_seconds']>=30],
           'before_over_30_invocation':[r for r in rows if r['before_invocation_seconds']!='' and r['before_invocation_seconds']>=30],
           'after_max_test':max(new.values(),key=lambda r:r['elapsed']),
           'after_max_invocation':max(invocation(r) for r in new.values()), 'observations':observations,
           'unchanged_production_and_fixture_inputs':sum(not p.endswith('_test.go') for p in before['inputs'])}
(review/'audit.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
