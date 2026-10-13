"""Credit hidden bytes once to their smallest blocking cause in both reports."""
import collections
import gzip
import json
from pathlib import Path

here = Path(__file__).resolve().parent
old = here.parent

def read_rows(path):
    with gzip.open(path, 'rt') as stream:
        return [json.loads(line) for line in stream]

def family(kind, reason):
    # Fold diagnostic payload types and names, preserving the diagnostic kind.
    prefixes = [
        'a function returning ', 'a value of type ', 'a field of type ',
        'reading ', 'assigning to ', 'storing ', 'replacing ',
        'a BinaryExpression with ', 'a PrefixUnaryExpression on ',
        'a method read as a value ', 'an unproven relation from ',
        'an arbitrary number or a value from another enum assigned to ',
        'an unproven value assigned to a numeric literal or enum member slot ',
        'optional property ',
    ]
    if kind == 'checker':
        return 'checker: diagnosed body'
    if kind == 'checker_dependency':
        return 'checker: diagnosed dependency body'
    for prefix in prefixes:
        if reason.startswith(prefix):
            if kind == 'Refused' and prefix == 'a value of type ':
                return 'Refused: invariant mutable relation'
            return kind + ': ' + prefix.strip() + ' <payload>'
    return kind + ': ' + reason

def attribute(report, rows, stock):
    root = Path(report['closure']['root'])
    causes = collections.defaultdict(list)
    texts = {f['text']: (f['kind'], f['reason']) for row in rows[1:]
             for f in row['findings'] if f['kind'] != 'Boundary'}
    def name(where):
        return Path(where.rsplit(':', 2)[0]).relative_to(root).as_posix()
    for row in rows[1:]:
        for finding in row['findings']:
            if finding['kind'] not in ('Boundary', 'SkippedDependency'):
                continue
            file = name(finding['where'])
            if file not in stock:
                continue
            if finding['kind'] == 'SkippedDependency':
                where = file + ':' + ':'.join(finding['where'].rsplit(':', 2)[1:])
                body = next(b for b in stock[file]['bodies'] if b['where'] == where)
                start, end = body['start'], body['end']
                label = family('checker_dependency', '')
            else:
                start, end = finding.get('start', 0), finding['end']
                kind, reason = texts.get(finding['text'], ('error', finding['text']))
                label = family(kind, reason)
            causes[file].append((end-start, start, end, label))
        for unit in row['units']:
            if unit['status'] in ('split_checker_body', 'skipped_checker_body'):
                file = name(unit['where'])
                start, end = unit['body_start'], unit['body_end']
                causes[file].append((end-start, start, end, family('checker', '')))
    totals = collections.Counter()
    for file, data in report['hidden']['files'].items():
        remaining = bytearray(data['bytes'])
        for start, end in data['hidden_ranges']:
            remaining[start:end] = b'\x01' * (end-start)
        # Shortest cause wins overlaps. Ties: start, end, family, lexical order.
        for _, start, end, label in sorted(set(causes[file])):
            amount = remaining[start:end].count(1)
            if amount:
                totals[label] += amount
                remaining[start:end] = b'\x00' * (end-start)
        assert not any(remaining), ('unattributed hidden byte', file)
    assert sum(totals.values()) == report['hidden']['hidden_bytes']
    return totals

before = json.loads((old / 'RESULT.json').read_text())
after = json.loads((here / 'RESULT.json').read_text())
def catalogue(base):
    with gzip.open(base / 'evidence/stock.json.gz', 'rt') as stream:
        return json.load(stream)
a = attribute(before, read_rows(old / 'evidence/full.jsonl.gz'), catalogue(old))
b = attribute(after, read_rows(here / 'evidence/full.jsonl.gz'), catalogue(here))
rows = [dict(family=k, before=a[k], after=b[k], delta=b[k]-a[k]) for k in a.keys() | b.keys()]
rows.sort(key=lambda row: (-abs(row['delta']), row['family']))
result = dict(definition='Exclusive hidden-byte attribution: smallest blocking span wins; ties by start, end and lexical family. Diagnostic kind retained, payload types/names folded by explicit prefixes in compare.py. Descriptive attribution, not isolated causal credit to commits.',
              before_total=sum(a.values()), after_total=sum(b.values()), families=rows)
(here / 'MOVERS.json').write_text(json.dumps(result, indent=2) + '\n')
for row in rows[:20]:
    print(f"{row['family']}: {row['before']:,} -> {row['after']:,} ({row['delta']:+,})")
print('PASS: exclusive credits sum to each hidden-byte headline')
