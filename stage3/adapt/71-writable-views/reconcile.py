#!/usr/bin/env python3
"""Reconcile all adaptation rows against complete before/after census evidence."""
import collections
import json
import pathlib
import sys

before_tree, before_raw, after_tree, after_raw, output = map(pathlib.Path, sys.argv[1:])
ledger = json.loads((pathlib.Path(__file__).parent / 'evidence/table-adaptations.json').read_text())

def read(tree, path):
    rows = [json.loads(line) for line in path.read_text().splitlines()]
    assert rows[0]['checker_rejected']
    files = {str(p.relative_to(tree)) for p in (tree / 'src/compiler').rglob('*.ts')}
    measured = [str(pathlib.Path(r['file']).relative_to(tree)) for r in rows[1:]]
    assert len(measured) == len(files) and set(measured) == files, 'incomplete or duplicate census'
    sites = {}
    for row in rows[1:]:
        for finding in row['findings']:
            if finding['kind'] != 'Refused':
                continue
            item = {key: value.replace(str(tree) + '/', '') if isinstance(value, str) else value
                    for key, value in finding.items()}
            key = tuple(item[k] for k in ('kind', 'where', 'reason', 'text'))
            sites[key] = item
    return rows[0], sites

before_meta, before = read(before_tree, before_raw)
after_meta, after = read(after_tree, after_raw)
first = collections.Counter(f['reason'] for f in before.values())
last = collections.Counter(f['reason'] for f in after.values())
records = []
for row in ledger['rows']:
    reason = row['reason']
    before_count, after_count = first[reason], last[reason]
    if reason.startswith('a value of type Node seen as Mutable<Node>, whose readonly field pos'):
        status = 'retained; flags are genuinely written, but a narrower view preserving readonly pos has not been audited'
    elif not row['writable_value_view']:
        status = 'outside writable-value-view implementation; type-only feasibility not audited'
    elif not before_count:
        status = 'not observed on current main; historical table count is not a current site count'
    elif not after_count:
        status = 'removed on current main'
    elif after_count < before_count:
        status = 'partially adapted; remaining sites retained'
    else:
        status = 'retained; no completed no-write or truthful-source proof'
    records.append(dict(row, main_before=before_count, main_after=after_count, status=status,
                        sites=[dict(where=f['where'], unit=f['unit'], phase=f['phase'])
                               for f in after.values() if f['reason'] == reason]))
result = dict(table_commit=ledger['commit'], measurement=before_meta['measurement'],
              count_definition='unique (kind, where, reason, text), all phases; exact table reason matching',
              table_reasons=ledger['reasons'], table_observations=ledger['observations'],
              current_before=sum(first[r['reason']] for r in ledger['rows']),
              current_after=sum(last[r['reason']] for r in ledger['rows']),
              removed=[before[k] for k in sorted(before.keys() - after.keys())],
              added=[after[k] for k in sorted(after.keys() - before.keys())],
              rows=records)
output.write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({k: result[k] for k in ('table_reasons', 'table_observations', 'current_before', 'current_after')}, indent=2))
