"""Reconcile scheduling evidence against the complete raw trace and kill artifact mutants."""
import collections
import copy
import gzip
import json
from pathlib import Path

BUCKET = Path(__file__).resolve().parent
trace = json.loads(gzip.decompress((BUCKET / 'trace.json.gz').read_bytes()))
ledger = json.loads((BUCKET / 'ledger.json').read_text())
status = json.loads((BUCKET / 'status.json').read_text())
end = {e['file']: index for index, e in enumerate(trace['events']) if e['event'] == 'end'}
raw = [r for r in trace['reads'] if r['statement'].startswith('src/compiler/') and '.generated.ts:' not in r['statement']]

def check(document):
    actual = collections.Counter()
    sites = set()
    for row in document['statements']:
        for binding in row['bindings']:
            if 'occurrences' not in binding:
                continue
            key = (row['statement'], binding['binding'], binding['provider'], binding['declaration'])
            actual[key] += binding['occurrences']
            assert binding['providerEndEvent'] == end[binding['provider']]
            assert binding['providerEndEvent'] < binding['firstReadEvent'] <= binding['lastReadEvent']
            sites.update((row['statement'], loc, binding['binding']) for loc in binding['readLocations'])
    expected = collections.Counter((r['statement'], r['binding'], r['provider'], r['declaration']) for r in raw)
    assert actual == expected
    assert sites == {(r['statement'], r['location'], r['binding']) for r in raw}
    assert sum(actual.values()) == document['originalReadOccurrences']
    assert len(sites) == document['originalReadSites']

check(ledger)
for label, mutate in [
    ('removed ledger row', lambda d: d['statements'].pop(0)),
    ('late provider end', lambda d: d['statements'][0]['bindings'][0].update(providerEndEvent=999999)),
]:
    mutant = copy.deepcopy(ledger)
    mutate(mutant)
    try:
        check(mutant)
    except AssertionError:
        print(label + ': caught')
    else:
        raise AssertionError(label + ': survived')
assert len(status) == 10
assert {r['file'] for r in status} == {f.relative_to(BUCKET).as_posix() for f in BUCKET.glob('*/main.a')}
for row in status:
    assert set(row) == {'file', 'tsc', 'reason', 'node', 'stage0'}
    assert set(row['node']) == {'stdout', 'stderr', 'exit'}
    assert set(row['stage0']) == {'outcome', 'what'}
    assert row['stage0']['outcome'] == 'Refused' and 'refuses an import cycle' in row['stage0']['what']
    assert row['node']['exit'] == (70 if '06_import_order_mutant' in row['file'] else 0)
    assert row['reason'] == 'an import cycle'
    assert (BUCKET / row['file']).read_text().startswith('// From TypeScript 6.0.3, src/compiler/')
print('Trace, ledger, ten fixture records and chronology: pass')
