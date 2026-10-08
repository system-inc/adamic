"""Rank the lazy owner's candidate dictionary table without inventing reachability."""
import gzip
import hashlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
HERE = Path(__file__).resolve().parent
SOURCE = ROOT / 'stage3/interface-downcasts/lazy/census/read-demand-pairs.json.gz'


def artifacts():
    inventory = json.loads(gzip.decompress(SOURCE.read_bytes()))
    pairs = [row for row in inventory if any(f in ['dictionary contracts', 'dictionary reads'] for f in row['families'])]
    assert len(pairs) == 29 and sum(row['read_count'] for row in pairs) == 228
    assert len({(row['type_id'], row['field']) for row in pairs}) == len(pairs)
    pairs.sort(key=lambda row: (-row['read_count'], row['type_id'], row['field']))
    completed_rows = json.loads((HERE / 'completed-read-pairs.json').read_text())
    completed = {(row['type_id'], row['field']): row for row in completed_rows}
    assert len(completed) == len(completed_rows)
    assert set(completed) <= {(row['type_id'], row['field']) for row in pairs}
    for row in completed_rows:
        assert all((HERE / fixture).is_file() for fixture in row['fixtures'])
    completed_pairs = [row for row in pairs if (row['type_id'], row['field']) in completed]
    completed_reads = sum(row['read_count'] for row in completed_pairs)
    for rank, pair in enumerate(pairs, 1):
        pair['rank'] = rank
        pair['source_integration'] = 'checked source reads' if (pair['type_id'], pair['field']) in completed else 'pending'
    summary = {
        'integration_base': 'ba59427ccc7afecae29a305c41e6e9c7867e5610',
        'input_sha256': hashlib.sha256(SOURCE.read_bytes()).hexdigest(),
        'measurement': 'Static candidate reads. Exact runtime reachability is unmeasured in the cited lazy reports.',
        'exact_reachable_pairs': None, 'exact_reachable_reads': None,
        'candidate_pairs': len(pairs), 'candidate_reads': sum(row['read_count'] for row in pairs),
        'completed_candidate_pairs': len(completed_pairs), 'completed_candidate_reads': completed_reads,
        'remaining_candidate_pairs': len(pairs) - len(completed_pairs), 'remaining_candidate_reads': sum(row['read_count'] for row in pairs) - completed_reads,
        'dictionary_contract_pairs': sum('dictionary contracts' in row['families'] for row in pairs),
        'dictionary_contract_reads': sum(row['read_count'] for row in pairs if 'dictionary contracts' in row['families']),
        'dictionary_read_pairs': sum('dictionary reads' in row['families'] for row in pairs),
        'dictionary_read_sites': sum(row['read_count'] for row in pairs if 'dictionary reads' in row['families']),
    }
    return {'candidate-pair-progress.json': pairs, 'candidate-summary.json': summary}


check = sys.argv[1:] == ['--check']
assert not sys.argv[1:] or check, 'usage: rank-lazy-demand.py [--check]'
for name, value in artifacts().items():
    contents = json.dumps(value, indent=2) + '\n'
    path = HERE / name
    if check:
        assert path.read_text() == contents, f'stale artifact: {name}'
    else:
        path.write_text(contents)
summary = artifacts()['candidate-summary.json']
print(f"{summary['remaining_candidate_pairs']} candidate pairs / {summary['remaining_candidate_reads']} candidate reads remain; exact runtime reachability unmeasured")
