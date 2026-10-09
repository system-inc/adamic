#!/usr/bin/env python3
"""Reject incomplete process maps, then sum each site's evaluation counts."""
import argparse
import json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument('sites', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('groups', nargs='+', help='name:expected-processes:directory')
args = parser.parse_args()
sites = json.loads(args.sites.read_text())
assert len(sites) == 47
ids = {r['id'] for r in sites}
assert len(ids) == 47
counts = {}
processes = {}
for group in args.groups:
    name, expected, directory = group.split(':', 2)
    files = sorted(Path(directory).glob('*.json'))
    assert len(files) == int(expected), f'{name}: expected {expected} exit maps, got {len(files)}'
    total = dict.fromkeys(ids, 0)
    for file in files:
        data = json.loads(file.read_text())
        assert set(data) == ids, f'{file}: missing or unknown site ids'
        assert all(type(n) is int and n >= 0 for n in data.values()), f'{file}: invalid counter'
        for key, n in data.items():
            total[key] += n
    counts[name] = total
    processes[name] = len(files)
rows = []
for r in sites:
    c = {name: counts[name][r['id']] for name in counts}
    rows.append({'id': r['id'], **c, 'total': sum(c.values()), 'bucket': 'rest' if sum(c.values()) else 'b'})
result = {'processes': processes, 'ran': sum(r['total'] > 0 for r in rows), 'unobserved': sum(r['total'] == 0 for r in rows), 'rows': rows}
args.output.write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({'processes': processes, 'ran': result['ran'], 'unobserved': result['unobserved']}))
