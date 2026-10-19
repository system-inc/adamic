#!/usr/bin/env python3
"""Compare saved CLI-backed census observations, using the original census IDs."""
import gzip
import hashlib
import json
from pathlib import Path
import sys

DIRECTORY = Path(__file__).resolve().parent
ARTIFACTS = DIRECTORY / 'optional-contracts-wave1'


def identity(row):
    return tuple(row[key] for key in ('file', 'line', 'column', 'code', 'message'))


def verify(directory=ARTIFACTS):
    table = json.loads((DIRECTORY / 'checker-stops-2026-10-19.json').read_text())
    records = {identity(row): row['id'] for row in table['diagnostics']}
    observations = {}
    for phase in ('before', 'after', 'mutant'):
        observations[phase] = json.loads(gzip.decompress((directory / f'{phase}.json.gz').read_bytes()))
        for run in observations[phase]['runs']:
            log = gzip.decompress((directory / f'{phase}-{run["name"]}.log.gz').read_bytes())
            assert hashlib.sha256(log).hexdigest() == run['logSha256'], (phase, run['name'], 'CLI bytes')
            assert run['cliApiByteIdentical'] is True
    before, after, mutant = (observations[phase] for phase in ('before', 'after', 'mutant'))
    assert before['roots'] == after['roots'] == mutant['roots']
    assert len(before['roots']) == 79
    changes = []
    assert len(before['runs']) == len(after['runs']) == len(mutant['runs']) == 6
    for old, new, reverted in zip(before['runs'], after['runs'], mutant['runs']):
        assert old['name'] == new['name'] == reverted['name']
        assert old['options'] == new['options'] == reverted['options']
        reference = next(run for run in table['runs'] if run['name'] == old['name'])
        assert sorted(records[identity(row)] for row in old['diagnostics']) == sorted(reference['diagnosticIds'])
        assert old['diagnostics'] == reverted['diagnostics'], (old['name'], 'revert did not restore diagnostics')
        old_rows = {identity(row): row for row in old['diagnostics']}
        new_rows = {identity(row): row for row in new['diagnostics']}
        removed = sorted(records[key] for key in old_rows.keys() - new_rows.keys())
        added = sorted(new_rows.keys() - old_rows.keys())
        assert removed == ['D4012d5810140'], (old['name'], removed)
        assert not added, (old['name'], added)
        for key in old_rows.keys() & new_rows.keys():
            assert old_rows[key] == new_rows[key], (old['name'], 'related information changed')
        old_hashes = {row['file']: row['sha256'] for row in old['sourceFiles']}
        new_hashes = {row['file']: row['sha256'] for row in new['sourceFiles']}
        mutant_hashes = {row['file']: row['sha256'] for row in reverted['sourceFiles']}
        assert old_hashes == mutant_hashes, (old['name'], 'mutant source bytes')
        assert old_hashes.keys() == new_hashes.keys()
        assert [file for file in old_hashes if old_hashes[file] != new_hashes[file]] == ['src/compiler/types.ts']
        changes.append(dict(run=old['name'], before=len(old_rows), after=len(new_rows),
                            mutant=len(reverted['diagnostics']), removed=removed, added=[]))
    return changes


if __name__ == '__main__':
    print(json.dumps(verify(Path(sys.argv[1]) if len(sys.argv) > 1 else ARTIFACTS), indent=2))
