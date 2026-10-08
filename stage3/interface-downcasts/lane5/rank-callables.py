#!/usr/bin/env python3
"""Rank the pinned potential callable reads without claiming certification."""
import json
from pathlib import Path

root = Path(__file__).resolve().parents[3]
source = root / 'stage3/interface-downcasts/lane2/read-census/pairs.json'
pairs = [pair for pair in json.loads(source.read_text())
         if 'callable contracts' in pair['families']]
pairs.sort(key=lambda pair: (-pair['read_count'], pair['type_id'], pair['field']))
assert len(pairs) == 308
assert sum(pair['read_count'] for pair in pairs) == 1503
assert all(pair['read_count'] == len(pair['sites']) for pair in pairs)
ranked = []
shapes = {}
for rank, pair in enumerate(pairs, 1):
    ranked.append(dict(pair, rank=rank, certified=False))
    shape = shapes.setdefault(pair['declared_type'], {'pairs': 0, 'reads': 0})
    shape['pairs'] += 1
    shape['reads'] += pair['read_count']
destination = Path(__file__).resolve().parent
(destination / 'callable-pairs-ranked.json').write_text(json.dumps(ranked, indent=2) + '\n')
ordered = sorted(shapes.items(), key=lambda item: (-item[1]['reads'], item[0]))
(destination / 'callable-shapes-ranked.json').write_text(json.dumps([
    dict(shape=shape, **counts) for shape, counts in ordered
], indent=2) + '\n')
print(f'{len(pairs)} pairs, 1503 potential reads, {len(shapes)} declared shapes; 0 certified')
for pair in ranked[:10]:
    print(f"{pair['rank']}: {pair['read_count']} reads {pair['type']}.{pair['field']}: {pair['declared_type']}")
