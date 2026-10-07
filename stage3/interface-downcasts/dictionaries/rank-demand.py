"""Preserve the frozen dictionary family; rank demand without claiming support."""
import collections
import gzip
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
SOURCE = ROOT / 'stage3/interface-downcasts/lane4/read-demand-pairs.json.gz'
DESTINATION = Path(__file__).resolve().parent


def inventory():
    pairs = [row for row in json.loads(gzip.decompress(SOURCE.read_bytes()))
             if 'dictionary contracts' in row['families']]
    assert len(pairs) == 409, len(pairs)
    assert sum(row['reads'] for row in pairs) == 2256
    assert len({(row['receiver_type_id'], row['field']) for row in pairs}) == 409
    pairs.sort(key=lambda row: (-row['reads'], row['receiver_type_id'], row['field']))
    shapes = collections.defaultdict(lambda: {'pairs': 0, 'reads': 0})
    for rank, row in enumerate(pairs, 1):
        row['rank'] = rank
        row['status'] = 'pending source integration'
        # Display evidence only. This is neither checker classification nor reassignment.
        row['node_array_display'] = any('NodeArray<' in t for t in row['declared_types'])
        shape = ' / '.join(row['declared_types'])
        shapes[shape]['pairs'] += 1
        shapes[shape]['reads'] += row['reads']
    summary = {
        'source_sha256': hashlib.sha256(SOURCE.read_bytes()).hexdigest(),
        'measurement': 'Conservative Unknown-fallback demand, not proven reaching-view counts',
        'pairs': len(pairs), 'reads': sum(row['reads'] for row in pairs),
        'completed_pairs': 0, 'completed_reads': 0,
        'remaining_pairs': len(pairs), 'remaining_reads': sum(row['reads'] for row in pairs),
        'node_array_display_pairs': sum(row['node_array_display'] for row in pairs),
        'node_array_display_reads': sum(row['reads'] for row in pairs if row['node_array_display']),
        'shapes': sorted(({'shape': shape, **counts} for shape, counts in shapes.items()),
                         key=lambda row: (-row['reads'], row['shape'])),
    }
    return {'pair-progress.json': pairs, 'demand-summary.json': summary}


if __name__ == '__main__':
    import sys
    check = sys.argv[1:] == ['--check']
    assert not sys.argv[1:] or check, 'usage: rank-demand.py [--check]'
    for name, value in inventory().items():
        contents = json.dumps(value, indent=2) + '\n'
        destination = DESTINATION / name
        if check:
            assert destination.read_text() == contents, f'stale ranking: {name}'
        else:
            destination.write_text(contents)
    print('409 pairs, 2256 reads; completed 0; remaining 409 pairs, 2256 reads')
