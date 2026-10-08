#!/usr/bin/env python3
"""Freeze this lane's demand without treating descriptors as admitted reads."""
import collections
import gzip
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parents[3]
source = root / 'stage3/interface-downcasts/lane4/read-demand-pairs.json.gz'
rows = json.loads(gzip.decompress(source.read_bytes()))
pairs = sorted((r for r in rows if 'object plus primitive union' in r['families']),
               key=lambda r: (-r['reads'], r['receiver_type_id'], r['field']))
assert len(pairs) == 73 and sum(r['reads'] for r in pairs) == 267
assert len({(r['receiver_type_id'], r['field']) for r in pairs}) == len(pairs)
shapes = collections.defaultdict(list)
for rank, row in enumerate(pairs, 1):
    row['rank'] = rank
    row['status'] = 'pending source integration'
    shapes[tuple(row['declared_types'])].append(row)
shape_rows = sorted(({'declared_types': list(shape), 'pairs': len(group),
                      'reads': sum(r['reads'] for r in group),
                      'pair_ranks': [r['rank'] for r in group]}
                     for shape, group in shapes.items()), key=lambda r: -r['reads'])
result = {'source_sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
          'assigned_pairs': len(pairs), 'assigned_reads': sum(r['reads'] for r in pairs),
          'completed_pairs': 0, 'completed_reads': 0, 'pairs': pairs, 'shapes': shape_rows}
Path(__file__).with_name('pair-progress.json').write_text(json.dumps(result, indent=2) + '\n')
print('73 pairs, 267 reads; completed 0 pairs, 0 reads')
for shape in shape_rows:
    print(f"{shape['reads']:3} reads {shape['pairs']:2} pairs: {'; '.join(shape['declared_types'])}")
