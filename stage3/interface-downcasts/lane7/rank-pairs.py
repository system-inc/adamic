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
# Credit only independently certified original contracts, never inferred support.
certificate = json.loads((Path(__file__).parent / 'original/certification.json').read_text())
certified = {(item['type_id'], item['field']): item['read_count']
             for item in certificate['pairs'] if item['status'] == 'certified'}
assigned = certificate.get('assigned_identifier', {})
if assigned.get('status') == 'certified':
    certified[(assigned['receiver_type_id'], assigned['field'])] = assigned['reads']
for filename in ['ranked-certification.json', 'ranked-operands-certification.json',
                 'ranked-minus-certification.json']:
    document = json.loads((Path(__file__).parent / 'original' / filename).read_text())
    for item in document if isinstance(document, list) else [document]:
        if item.get('status') == 'certified':
            certified[(item['type_id'], item['field'])] = item['read_count']
completed_reads = 0
completed_pairs = 0
for rank, row in enumerate(rows, 1):
    row['rank'] = rank
    row['status'] = 'pending original runtime certificate'
    key = (row['receiver_type_id'], row['field'])
    if key in certified:
        assert certified[key] == row['reads'], ('read-count drift', key)
        row['status'] = 'certified original representative read contract'
        completed_pairs += 1
        completed_reads += row['reads']
    # Only this exact display split was assigned by the ruling. Other brands
    # require checker constituent evidence before their ownership is decided.
    row['owner'] = 'lane 4' if row['declared_types'] == ['__String'] else 'lane 7 queue, overlaps unresolved'
result = {'source_sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
          'measurement': 'Unknown-fallback family demand, not successful lowering',
          'pairs': len(rows), 'reads': sum(row['reads'] for row in rows),
          'completed_pairs': completed_pairs, 'completed_reads': completed_reads,
          'remaining_candidate_pairs': len(rows) - completed_pairs,
          'remaining_candidate_reads': sum(row['reads'] for row in rows) - completed_reads,
          'certification_scope': 'Representative original contracts; candidate inventory is conservative and includes other owners; exact runtime reachability unmeasured',
          'ranked_pairs': rows}
destination = Path(__file__).with_name('pair-progress.json')
destination.write_text(json.dumps(result, indent=2) + '\n')
print(f'198 pairs, 1145 reads; certified {completed_pairs} pairs, {completed_reads} reads; pending candidates {198-completed_pairs} pairs, {1145-completed_reads} reads')
print('Exact __String display: 30 pairs, 543 reads assigned to lane 4')
