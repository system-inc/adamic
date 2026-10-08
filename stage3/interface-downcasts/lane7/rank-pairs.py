#!/usr/bin/env python3
"""Preserve the frozen intersection queue without inferring runtime support."""
import gzip
import hashlib
import json
from pathlib import Path

root = Path(__file__).resolve().parents[3]
source = root / 'stage3/interface-downcasts/lane4/read-demand-pairs.json.gz'
rows = [row for row in json.loads(gzip.decompress(source.read_bytes()))
        if 'intersection field contract' in row['families']]
rows.sort(key=lambda row: (-row['reads'], row['receiver_type_id'], row['field']))
assert len(rows) == 198 and sum(row['reads'] for row in rows) == 1145
for rank, row in enumerate(rows, 1):
    row['rank'] = rank
    row['status'] = 'pending source integration'
    # Only this exact display split was assigned by the ruling. Other brands
    # require checker constituent evidence before their ownership is decided.
    row['owner'] = 'lane 4' if row['declared_types'] == ['__String'] else 'lane 7 queue, overlaps unresolved'
result = {'source_sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
          'measurement': 'Unknown-fallback family demand, not successful lowering',
          'pairs': len(rows), 'reads': sum(row['reads'] for row in rows),
          'completed_pairs': 0, 'completed_reads': 0, 'ranked_pairs': rows}
destination = Path(__file__).with_name('pair-progress.json')
destination.write_text(json.dumps(result, indent=2) + '\n')
print('198 pairs, 1145 reads; completed 0; remaining 198 pairs, 1145 reads')
print('Exact __String display: 30 pairs, 543 reads assigned to lane 4')
