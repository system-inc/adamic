#!/usr/bin/env python3
"""Compare actual runner outcomes without changing classification or verdicts."""
import collections
import csv
import json
import pathlib
import sys

root = pathlib.Path(__file__).resolve().parent
before = [json.loads(line) for line in (root.parent / 'regexp-test262-runner/results.jsonl').read_text().splitlines()]
after = [json.loads(line) for line in pathlib.Path(sys.argv[1]).read_text().splitlines()]
assert len(before) == len(after) == 1879
assert {row['path'] for row in before} == {row['path'] for row in after}
old = {row['path']: row for row in before}
new = {row['path']: row for row in after}


def group(row):
    reason = row.get('reason', '')
    if row['kind'] == 'pass': return 'Pass'
    if row['kind'] not in ('refused', 'skipped'): return row['kind']
    if reason.startswith('error TS'): return 'Stock TypeScript rejection (audit separately)'
    if 'constructor-identity assertion' in reason: return 'Runner constructor guard (native and adapted Node succeeded)'
    if 'a RegExp constructor that throws SyntaxError' in reason: return 'Constant constructor SyntaxError not implemented'
    if 'V8 diagnostic wording' in reason: return 'Observable SyntaxError diagnostic refused'
    if 'nonconstant pattern' in reason: return 'Runtime pattern explicitly refused'
    if 'nonconstant flags' in reason: return 'Runtime flags explicitly refused'
    if row['kind'] == 'skipped' and reason.startswith('feature Symbol.'): return 'Symbol protocol/species classifier skip'
    if row['kind'] == 'skipped': return reason
    return 'Other compiler refusal'


with (root / 'before-after.tsv').open('w') as output:
    writer = csv.writer(output, delimiter='\t')
    writer.writerow(['outcome', 'before', 'after'])
    for name in ('pass', 'fail', 'refused', 'crashed', 'skipped'):
        writer.writerow([name, sum(row['kind'] == name for row in before), sum(row['kind'] == name for row in after)])
    writer.writerow(['total', len(before), len(after)])

with (root / 'reason-groups.tsv').open('w') as output:
    writer = csv.writer(output, delimiter='\t')
    writer.writerow(['reason', 'before', 'after'])
    b, a = collections.Counter(map(group, before)), collections.Counter(map(group, after))
    for reason in sorted(b.keys() | a.keys(), key=lambda reason: (-max(a[reason], b[reason]), reason)):
        writer.writerow([reason, b[reason], a[reason]])

with (root / 'transitions.tsv').open('w') as output:
    writer = csv.writer(output, delimiter='\t')
    writer.writerow(['path', 'before', 'after', 'before reason', 'after reason'])
    for path in sorted(old):
        b, a = old[path], new[path]
        if (b['kind'], b.get('reason', '')) != (a['kind'], a.get('reason', '')):
            writer.writerow([path, b['kind'], a['kind'], b.get('reason', ''), a.get('reason', '')])

with (root / 'directories.tsv').open('w') as output:
    writer = csv.writer(output, delimiter='\t')
    writer.writerow(['directory', 'before pass', 'after pass', 'before refused', 'after refused', 'before skipped', 'after skipped', 'after fail', 'after crashed', 'total'])
    for directory in sorted({row['directory'] for row in before}):
        b = collections.Counter(row['kind'] for row in before if row['directory'] == directory)
        a = collections.Counter(row['kind'] for row in after if row['directory'] == directory)
        writer.writerow([directory, b['pass'], a['pass'], b['refused'], a['refused'], b['skipped'], a['skipped'], a['fail'], a['crashed'], sum(a.values())])

with (root / 'all-reasons.tsv').open('w') as output:
    writer = csv.writer(output, delimiter='\t')
    writer.writerow(['outcome', 'reason', 'before', 'after'])
    b = collections.Counter((row['kind'], row.get('reason', '')) for row in before)
    a = collections.Counter((row['kind'], row.get('reason', '')) for row in after)
    for key in sorted(b.keys() | a.keys()): writer.writerow([*key, b[key], a[key]])

regressions = [path for path in old if old[path]['kind'] == 'pass' and new[path]['kind'] != 'pass']
disagreements = [row for row in after if row['kind'] in ('fail', 'crashed')]
print(json.dumps({'before': collections.Counter(row['kind'] for row in before), 'after': collections.Counter(row['kind'] for row in after), 'regressions': regressions, 'disagreements': disagreements}, indent=2))
assert not regressions and not disagreements
