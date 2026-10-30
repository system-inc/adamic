#!/usr/bin/env python3
"""Compare complete first-error latent runs using the meter's distinct-site key."""
import collections
import hashlib
import json
from pathlib import Path
import sys

before, after, output = map(Path, sys.argv[1:])
def read(path):
    records = [json.loads(line) for line in path.read_text().splitlines()]
    assert records[0]['checker_rejected']
    roots = {record['file'].split('/src/compiler/')[0] for record in records[1:]}
    assert len(roots) == 1
    root = next(iter(roots)) + '/'
    normalize = lambda value: value.replace(root, '')
    sites = set()
    for record in records[1:]:
        for finding in record.get('findings') or []:
            if finding['kind'] in ('NotYet', 'Refused'):
                sites.add((finding['kind'], normalize(finding['where']), normalize(finding['reason']), normalize(finding['text'])))
    return records, sites
left, a = read(before)
right, b = read(after)
assert len(left) == len(right), 'same complete file coverage'
assert {r['file'].split('/src/compiler/')[1] for r in left[1:]} == {r['file'].split('/src/compiler/')[1] for r in right[1:]}
x = collections.Counter(s[2] for s in a)
y = collections.Counter(s[2] for s in b)
rows = []
for reason in sorted(x.keys() | y.keys()):
    if x[reason] != y[reason]:
        rows.append({'reason':reason, 'before':x[reason], 'after':y[reason], 'delta':y[reason]-x[reason],
                     'examples':[dict(zip(('kind','where','reason','text'),s)) for s in sorted(b-a) if s[2] == reason][:3]})
result = {'mode':'first-error', 'distinct_key':['kind','where','reason','text'], 'normalization':'remove each adapted tree root from embedded locations in where, reason and text', 'files':len(left)-1,
          'before_sites':len(a), 'after_sites':len(b), 'changed_reasons':rows,
          'new_or_growing':[r for r in rows if r['delta']>0],
          'raw_sha256':{'before':hashlib.sha256(before.read_bytes()).hexdigest(),'after':hashlib.sha256(after.read_bytes()).hexdigest()}}
output.mkdir(parents=True,exist_ok=True)
(output/'census-delta.json').write_text(json.dumps(result,indent=2)+'\n')
lines=['| Reason | Before | After | Change |','|---|---:|---:|---:|']
for row in rows:
    reason=row['reason'].replace('|','&#124;').replace('\n',' ')
    lines.append(f"| {reason} | {row['before']} | {row['after']} | {row['delta']:+} |")
(output/'census-delta.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({'files':len(left)-1,'before':len(a),'after':len(b),'changed_reasons':len(rows),'new_or_growing':len(result['new_or_growing'])}))
