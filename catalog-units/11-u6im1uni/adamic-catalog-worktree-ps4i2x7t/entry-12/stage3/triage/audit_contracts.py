#!/usr/bin/env python3
"""Compare every candidate ledger row and full chain with a completed census."""
import collections
import hashlib
import json
from pathlib import Path
import re
import sys

CODES = {2322, 2339, 2538, 2488, 2769, 18046, 2345}
PATTERN = re.compile(r'^(.*?):(\d+):(\d+): error TS(\d+): (.*)', re.S)


def candidates(path):
    records = [json.loads(line) for line in Path(path).read_text().splitlines()]
    assert len(records[-1]['roots']) > 1, 'census has no whole-program record'
    assert records[-1]['kind'] == 'checker', 'expected checker whole-program result'
    rows = []
    for diagnostic in records[-1]['diagnostics']:
        match = PATTERN.fullmatch(diagnostic)
        if match and int(match[4]) in CODES:
            file = 'src/compiler/' + match[1].split('/src/compiler/', 1)[1]
            rows.append(dict(file=file, line=int(match[2]), column=int(match[3]),
                             code=int(match[4]), chain=match[5]))
    return rows


def key(row):
    return tuple(row[name] for name in ['file', 'line', 'column', 'code', 'chain'])


def audit(census, ledger):
    expected = collections.Counter(map(key, candidates(census)))
    actual = collections.Counter(map(key, ledger['rows']))
    missing, extra = expected - actual, actual - expected
    assert not missing and not extra, f'dropped/changed rows: {list(missing.items())[:3]}; extra: {list(extra.items())[:3]}'
    counts = dict(sorted(collections.Counter(str(row['code']) for row in ledger['rows']).items()))
    assert ledger['candidate_counts'] == counts, 'incorrect candidate counts'
    assert len({row['id'] for row in ledger['rows']}) == len(ledger['rows']), 'duplicate row ID'
    assert ledger['census_sha256'] == hashlib.sha256(Path(census).read_bytes()).hexdigest(), 'wrong census bytes'
    if 'origin_review' in ledger:
        manifest = json.loads((Path(__file__).parent / 'evidence/origin-review.json').read_text())
        reviewed = [r for r in ledger['rows'] if r.get('origin_review_cohort') == 'original-434']
        assert len(reviewed) == len(manifest) == 434, 'dropped origin review'
        expected_review = {r['id']: r for r in manifest}
        for row in reviewed:
            proof = expected_review[row['id']]
            assert all(row[k] == proof[k] for k in ('file', 'line', 'column', 'origin_class')), 'changed origin classification'
            assert row['origin_trace']['id'] == row['id'] and row['origin_trace']['steps'], 'missing origin trace'
            assert row['class'] == row['origin_class'], 'wrong class group'
            assert not row['smallest_edit'], 'unverified new edit'
        counts_origin = collections.Counter(r['origin_class'] for r in reviewed)
        assert ledger['origin_review']['population'] == 434
        assert collections.Counter(ledger['origin_review']['counts']) == counts_origin, 'wrong origin counts'
        print(f'PASS: 434 origin locations, traces and classes: {dict(counts_origin)}')
    print(f'PASS: {len(ledger["rows"])} exact rows, chains, locations and counts: {counts}')


if __name__ == '__main__':
    if len(sys.argv) != 3:
        sys.exit('usage: audit_contracts.py census.jsonl contracts.json')
    audit(sys.argv[1], json.loads(Path(sys.argv[2]).read_text()))
