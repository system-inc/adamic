#!/usr/bin/env python3
"""Compare complete census runs without treating skipped work as removal."""
import collections
import gzip
import hashlib
import json
import pathlib
import sys

before_tree, before_raw, after_tree, after_raw, output = map(pathlib.Path, sys.argv[1:])
output.mkdir(exist_ok=False)

def read(tree, raw, name):
    raw_rows = [json.loads(line) for line in raw.read_text().splitlines()]
    raw_tree = raw_rows[1]["file"].split("/src/compiler/")[0]
    assert raw_tree and all(r["file"].startswith(raw_tree + "/src/compiler/") for r in raw_rows[1:])
    def normalize(value):
        if isinstance(value, str): return value.replace(raw_tree + '/', '')
        if isinstance(value, list): return [normalize(v) for v in value]
        if isinstance(value, dict): return {k: normalize(v) for k, v in value.items()}
        return value
    rows = [normalize(row) for row in raw_rows]
    assert rows[0]['checker_rejected']
    measured = [r['file'] for r in rows[1:]]
    files = {str(p.relative_to(tree)) for p in (tree / 'src/compiler').rglob('*.ts')}
    assert len(measured) == len(files) and set(measured) == files
    with gzip.open(output / (name + '.jsonl.gz'), 'wt') as stream:
        for row in rows: stream.write(json.dumps(row) + '\n')
    sites = {tuple(f[k] for k in ('kind', 'where', 'reason', 'text')): f
             for row in rows[1:] for f in row['findings']
             if f['kind'] == 'Refused' and ' seen as ' in f['reason']}
    eligible = {u['where'] for r in rows[1:] for u in r['units'] if u['status'] != 'skipped_checker_body'}
    hashes = {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest()
              for p in sorted((tree / 'src/compiler').rglob('*.ts'))}
    return rows[0], sites, eligible, hashes

bm, before, bu, bh = read(before_tree, before_raw, 'before')
am, after, au, ah = read(after_tree, after_raw, 'after')
result = dict(measurement=bm['measurement'],
    count_definition='unique (kind, where, reason, text), Refused reason contains seen as',
    before=len(before), after=len(after), removed=[before[k] for k in sorted(before.keys() - after.keys())],
    added=[after[k] for k in sorted(after.keys() - before.keys())],
    before_per_reason=dict(collections.Counter(f['reason'] for f in before.values())),
    after_per_reason=dict(collections.Counter(f['reason'] for f in after.values())),
    checker_diagnostics_before=len(bm['diagnostics']), checker_diagnostics_after=len(am['diagnostics']),
    checker_added=sorted(set(am['diagnostics'])-set(bm['diagnostics'])),
    checker_removed=sorted(set(bm['diagnostics'])-set(am['diagnostics'])),
    eligible_units_before=len(bu), eligible_units_after=len(au),
    newly_eligible_units=sorted(au-bu), no_longer_eligible_units=sorted(bu-au),
    source_hashes_before=bh, source_hashes_after=ah)
(output / 'census.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({k: result[k] for k in ['before', 'after', 'checker_diagnostics_before', 'checker_diagnostics_after',
    'newly_eligible_units', 'no_longer_eligible_units']}, indent=2))
