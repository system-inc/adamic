#!/usr/bin/env python3
"""Verify exact census identities and every wave-2 group disposition."""
import gzip
import hashlib
import json
from pathlib import Path
import sys

DIRECTORY = Path(__file__).resolve().parent
ARTIFACTS = DIRECTORY / 'optional-contracts-wave2'


def identity(row):
    return tuple(row[key] for key in ('file', 'line', 'column', 'code', 'message'))


def read(path):
    return json.loads(gzip.decompress(path.read_bytes()))


def compare_runs(old, new, ids):
    assert old['name'] == new['name'] and old['options'] == new['options']
    old_rows = {identity(row): row for row in old['diagnostics']}
    new_rows = {identity(row): row for row in new['diagnostics']}
    assert not new_rows.keys() - old_rows.keys(), (old['name'], 'new diagnostic chain')
    for key in new_rows.keys():
        assert old_rows[key] == new_rows[key], (old['name'], 'related-information change')
    return old_rows, new_rows, sorted(ids[key] for key in old_rows.keys() - new_rows.keys())


def verify(directory=ARTIFACTS):
    original = json.loads((DIRECTORY / 'checker-stops-2026-10-19.json').read_text())
    split = json.loads((DIRECTORY / 'optional-contracts-2026-10-19.json').read_text())
    dispositions = json.loads((directory / 'groups.json').read_text())
    ids = {identity(row): row['id'] for row in original['diagnostics']}
    before, after = (read(directory / f'{phase}.json.gz') for phase in ('before', 'after'))
    assert before['roots'] == after['roots'] and len(before['roots']) == 79
    assert len(before['runs']) == len(after['runs']) == 8
    results = []
    for phase, observed in [('before', before), ('after', after)]:
        for run in observed['runs']:
            log = gzip.decompress((directory / f'{phase}-{run["name"]}.log.gz').read_bytes())
            assert hashlib.sha256(log).hexdigest() == run['logSha256'], (phase, run['name'])
            assert run['cliApiByteIdentical'] is True
    for old, new in zip(before['runs'], after['runs']):
        old_rows, new_rows, removed = compare_runs(old, new, ids)
        expected = [] if old['name'] == 'stock-no-exact-control' else dispositions['removed_ids']
        assert removed == sorted(expected), (old['name'], removed, expected)
        results.append(dict(run=old['name'], before=len(old_rows), after=len(new_rows), removed_ids=removed, added_ids=[]))
    primary_removed = set(results[0]['removed_ids'])
    rows = [row for row in split['rows'] if row['reasonGroup'] not in ('optional-method', 'nullable-return') and not any(owner['publicApi'] for owner in row['owners'])]
    assert len(rows) == len(dispositions['rows']) == 30
    assert len(primary_removed) == 30
    repaired = {row['id'] for row in dispositions['rows'] if row['status'] == 'repaired'}
    declined = {row['id'] for row in dispositions['rows'] if row['status'] == 'declined'}
    assert len(repaired) == 24 and len(declined) == 6
    assert repaired <= primary_removed and not declined.intersection(primary_removed)
    assert repaired | declined == {row['id'] for row in rows}
    for row in split['rows']:
        if row['reasonGroup'] == 'optional-method' or any(owner['publicApi'] for owner in row['owners']):
            assert row['id'] not in primary_removed, row['id']
    for group in dispositions['groups']:
        members = [row for row in dispositions['rows'] if row['reasonGroup'] == group['reason']]
        assert group['count'] == len(members)
        assert sorted(group['removed_ids']) == sorted(row['id'] for row in members if row['status'] == 'repaired')
    for result_path in sorted((directory / 'mutants').glob('*/result.json')):
        result = json.loads(result_path.read_text())
        assert hashlib.sha256((result_path.parent / 'resolver-mutant.cjs').read_bytes()).hexdigest() == result['mutant_helper_sha256'], (result['group'], 'mutant resolver bytes')
        if result['group'] == 'nested-expression':
            assert all(witness['exit'] == 0 for witness in result['witnesses'])
        else:
            assert any(witness['exit'] == 1 for witness in result['witnesses']), result['group']
        observed = read(result_path.parent / 'census.json.gz')
        assert len(observed['runs']) == 6
        for run, fixed in zip(observed['runs'], after['runs'][:6]):
            log = gzip.decompress((result_path.parent / f'{run["name"]}.log.gz').read_bytes())
            assert hashlib.sha256(log).hexdigest() == run['logSha256']
            fixed_ids = {identity(row) for row in fixed['diagnostics']}
            mutant_ids = {identity(row) for row in run['diagnostics']}
            returned = {ids[key] for key in mutant_ids - fixed_ids if key in ids}
            if result['group'] == 'nested-expression':
                assert mutant_ids == fixed_ids
            elif result['group'] == 'public-blocker':
                assert len(mutant_ids - fixed_ids) == 3
            else:
                assert returned.intersection(result['group_ids_returned']), result['group']
    assert len(list((directory / 'mutants').glob('*/result.json'))) == 8
    return results


if __name__ == '__main__':
    print(json.dumps(verify(Path(sys.argv[1]) if len(sys.argv) > 1 else ARTIFACTS), indent=2))
