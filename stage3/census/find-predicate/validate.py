"""Validate archived evidence and report accounting independently of compare.py.

Usage: python3 validate.py [SHARED_ADAPTED_TREE]
"""
import collections
import copy
import gzip
import hashlib
import json
from pathlib import Path
import sys
import re

O = Path(__file__).resolve().parent
D = json.loads((O / 'delta.json').read_text())
E = O / 'evidence'

def read(name):
    with gzip.open(E / name, 'rt') as f:
        return [json.loads(line) for line in f if line.strip()]

def check(d):
    for name in ('raw_findings', 'distinct_reads'):
        transitions = d['find_sites'][name]['transitions']
        summary = d['find_sites'][name]['summary']
        for status in ('gone', 'moved', 'unchanged'):
            assert summary[status] == sum(t['status'] == status for t in transitions)
        assert sum(summary[s] for s in ('gone', 'moved', 'unchanged')) == len(transitions)
        for t in transitions:
            assert t['base']['reason'] in d['matched_reasons']
            if t['status'] == 'gone':
                assert not t['head']
            elif t['status'] == 'moved':
                assert t['head'] and all(f['reason'] != t['base']['reason'] for f in t['head'])
            else:
                assert any(f['reason'] == t['base']['reason'] for f in t['head'])
    changes = d['raw_findings']['unit_first_stop_delta']
    expected = sum(bool(r['base'] and r['head']) for r in changes)
    assert expected == sum(r['type'] == 'changed_first_lowering_stop' for r in d['regression_bar'])
    assert len(d['sort_cross_check']) == 25
    assert len({r['historical']['unit'] for r in d['sort_cross_check']}) == 25

check(D)
base, head = [read(f'{s}-latent.jsonl.gz') for s in ('base', 'head')]
assert base[0] == head[0] and base[0]['latent_mode'] == 'first-error'
assert {r['file'] for r in base[1:]} == {r['file'] for r in head[1:]}
for side, rows in [('base', base), ('head', head)]:
    fs = [f for r in rows[1:] for f in r['findings']]
    unique = {(f['kind'],f['where'],f['reason'],f['text']) for f in fs}
    counts = dict(collections.Counter(k[0] for k in unique))
    assert counts == D['raw_findings']['totals']['before' if side == 'base' else 'after']
    if side == 'base':
        direct = {x['where'] for x in D['find_context'] if x['direct_find']}
        selected = {(f['kind'], f['where'], f['reason'], f['text']) for f in fs
                    if re.search(r'overload|predicate', f['reason']) and
                    (re.search(r'\bfind\b', f['reason']) or f['where'] in direct)}
        actual = {(t['base']['kind'], t['base']['where'], t['base']['reason'], t['base']['text'])
                  for t in D['find_sites']['raw_findings']['transitions']}
        assert selected == actual
    times = read(f'{side}-times.jsonl.gz')
    units = {u['where']:u for r in rows[1:] for u in r['units']}
    assert {t['where'] for t in times} == {w for w,u in units.items() if u['status'] in ('attempted','panic')}
    assert len(times) == D['timing_sums'][side]['attempts']
    for k in ('wall_seconds', 'cpu_seconds'):
        assert sum(t[k] for t in times) == D['timing_sums'][side][k]
    assert D['resources'][side]['exit_code'] == 0
    assert not (E / f'{side}-status.txt').read_text()
    assert (E / f'{side}-memory-before.txt').read_bytes() == (E / f'{side}-memory-after.txt').read_bytes()

for name, expected in json.loads((O/'artifacts.sha256.json').read_text()).items():
    assert hashlib.sha256((O/name).read_bytes()).hexdigest() == expected, name
if len(sys.argv) == 2:
    A = Path(sys.argv[1])
    manifest = json.loads((E/'source-manifest-before.json').read_text())
    assert manifest == {str(p.relative_to(A)): hashlib.sha256(p.read_bytes()).hexdigest()
                        for p in sorted((A/'src').rglob('*')) if p.is_file()}
    for file, sha in D['source_hashes'].items():
        assert hashlib.sha256((A/file).read_bytes()).hexdigest() == sha
    assert {r['file'] for r in base[1:]} == {str(p.relative_to(A)) for p in (A/'src/compiler').rglob('*.ts')}
# An incorrect headline must be rejected by these same accounting assertions.
mutant = copy.deepcopy(D)
mutant['find_sites']['raw_findings']['summary']['gone'] += 1
try:
    check(mutant)
except AssertionError:
    print('Headline-count mutant rejected.')
else:
    raise AssertionError('headline-count mutant survived')
for p in O.rglob('*'):
    if p.is_file() and p.suffix in ('.md', '.py', '.sh'):
        assert '\u2014' not in p.read_text() and '\u2013' not in p.read_text(), p
print('Archived counts, coverage, transitions, regression bar, unit times, resources and artifact hashes validated.')
print('Shared source hashes validated.' if len(sys.argv) == 2 else 'Shared source check omitted; pass the adapted tree to include it.')
