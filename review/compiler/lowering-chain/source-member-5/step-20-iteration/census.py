"""Extract iteration findings without normalizing exact reasons or adding overlapping bytes."""
import argparse
import gzip
import json
from pathlib import Path


def belongs(reason):
    return (reason.startswith('for...of') or reason.startswith('a for...of')
            or reason in ('iterating a value', 'a generator function', 'yield (generators)',
                          'a YieldExpression as a statement')
            or 'iterator' in reason.lower() or 'iterable' in reason.lower())


def short(where):
    return where.split('/src/')[-1].removeprefix('compiler/')


def inventory(path):
    groups = {}
    with gzip.open(path, 'rt') if str(path).endswith('.gz') else open(path) as stream:
        for line in stream:
            row = json.loads(line)
            for finding in row.get('findings', []):
                if finding['kind'] not in ('Refused', 'NotYet') or not belongs(finding['reason']):
                    continue
                key = (finding['kind'], finding['reason'])
                group = groups.setdefault(key, {'roots': set(), 'sites': set(), 'phases': set()})
                group['roots'].add(short(finding['unit']))
                group['sites'].add(short(finding['where']))
                group['phases'].add(finding['phase'])
    return [dict(kind=kind, reason=reason, units=len(group['roots']),
                 roots=len(group['sites']), witnesses=sorted(group['sites'])[:3],
                 all_sites=sorted(group['sites']), root_units=sorted(group['roots']),
                 phases=sorted(group['phases']))
            for (kind, reason), group in sorted(groups.items())]


def audit(rows):
    for row in rows:
        assert row['roots'] == len(set(row['all_sites'])), 'root deduplication changed'
        assert row['units'] == len(set(row['root_units'])), 'unit deduplication changed'
        assert row['witnesses'] == sorted(set(row['all_sites']))[:3], 'witness selection changed'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('census', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('--compiler', required=True)
    parser.add_argument('--mutant', choices=('roots', 'units', 'witnesses', 'coverage'))
    args = parser.parse_args()
    with gzip.open(args.census, 'rt') if str(args.census).endswith('.gz') else open(args.census) as stream:
        records = [json.loads(line) for line in stream]
    if args.mutant == 'coverage':
        records.pop()
    header, files = records[0], records[1:]
    assert len(files) == len({row['file'] for row in files}), 'duplicate file record'
    assert {row['file'] for row in files} == set(header['sources']), 'resolved source coverage changed'
    rows = inventory(args.census)
    assert rows, 'no iteration findings selected'
    if args.mutant == 'roots':
        rows[0]['roots'] += 1
    elif args.mutant == 'units':
        rows[0]['units'] += 1
    elif args.mutant == 'witnesses':
        rows[0]['witnesses'] = ['invented.a:1:1']
    audit(rows)
    args.output.write_text(json.dumps({'compiler': args.compiler, 'input': str(args.census),
                                     'rows': rows}, indent=2) + '\n')
    print(f'{len(rows)} exact reasons audited')


if __name__ == '__main__':
    main()
