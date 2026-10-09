"""Audit exact diagnostic retirements, keeping other unit boundaries visible."""
import argparse
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('archive', type=Path)
    parser.add_argument('--mutant', choices=('coverage', 'retired', 'witness', 'declaration'))
    args = parser.parse_args()
    with gzip.open(args.archive, 'rt') as stream:
        evidence = json.load(stream)
    rows = evidence['rows']
    if args.mutant == 'coverage':
        rows.pop()
    if args.mutant == 'retired':
        rows[0]['retired'] = not rows[0]['retired']
    if args.mutant == 'witness':
        rows[0]['where'] = 'invented.a:1:1'
    if args.mutant == 'declaration':
        next(iter(evidence['declaration_snapshots'].values())).pop()
    for digest, declarations in evidence['declaration_snapshots'].items():
        assert hashlib.sha256(json.dumps(declarations, sort_keys=True, separators=(',', ':')).encode()).hexdigest() == digest, 'declaration snapshot changed'
    expected = {r['where'] for r in evidence['targets']}
    assert len(rows) == len({r['where'] for r in rows}) == len(expected), 'retirement coverage changed'
    assert {r['where'] for r in rows} == expected, 'retirement witness changed'
    count = 0
    for row in rows:
        for phase in ('before', 'after'):
            assert row[phase]['declarations_sha256'] in evidence['declaration_snapshots'], 'missing declaration snapshot'
        signature = '/tmp/scout-adapted/src/compiler/' + row['where']
        def matches(record):
            return any(f['where'] == signature and f['kind'] == 'NotYet'
                       and f['reason'] == 'for...of over an object'
                       and f['phase'] == 'lowering' for f in (record.get('findings') or []))
        reproduced = matches(row['before'])
        assert row['reproduced_before'] == reproduced, 'base reproduction classification changed'
        actual = reproduced and not matches(row['after'])
        assert row['retired'] == actual, 'retirement classification changed'
        assert row['before']['measurement'] == row['after']['measurement'] == 'measured on a checker-rejected entry-root program', 'measurement label changed'
        assert row['before']['units'] == row['after']['units'], 'selected attempted unit changed'
        count += actual
    assert evidence['retired_roots'] == count, 'retirement total changed'
    print(f'{len(rows)} exact roots audited; {count} old object-iteration signatures retired')
    print('base signatures not reproduced:', [r['where'] for r in rows if not r['reproduced_before']])
    remaining = Counter(r['type'].split('<')[0] for r in rows if r['reproduced_before'] and not r['retired'])
    print('remaining original signatures:', dict(remaining))


if __name__ == '__main__':
    main()
