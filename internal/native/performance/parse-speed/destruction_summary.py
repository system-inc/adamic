#!/usr/bin/env python3
"""Reconcile destruction counters with per-type Callgrind clones; reject mismatches."""
import argparse
import collections
import json
from pathlib import Path
import re

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('directory', type=Path)
args = parser.parse_args()
base = args.directory
text = (base / 'destruction-counter.stderr').read_text()
labels = ['null', 'string', 'object', 'array_reference', 'map', 'cell', 'closure', 'map_iterator', 'number', 'boolean', 'weak', 'node', 'array_number']
counts = collections.Counter()
phases = collections.Counter()
bytes_at_death = collections.Counter()
references = {}
allocated = 0
ends = []
for line in text.splitlines():
    values = {k: int(v) for k, v in re.findall(r'(\w+)=(\d+)', line)}
    if line.startswith('destroy '):
        counts[values['type']] += values['count']
        phases[values['phase'], values['type']] += values['count']
        bytes_at_death[values['type']] += values['bytes']
    elif line.startswith('references '):
        references[values['type'], values['state']] = values
    elif line.startswith('allocate '):
        allocated += values['count']
    elif line.startswith('file_end '):
        ends.append(values)
    elif line.startswith('nodes '):
        nodes = values
assert len(ends) == 77 and all(row['live_nodes'] == 0 for row in ends), 'trees survive file cleanup'
assert allocated == sum(counts.values()), 'allocated and destroyed counts differ'
assert nodes['allocated'] == counts[11] and nodes['live'] == 0, 'node balance differs'
assert counts[11] == phases[3, 11], 'nodes destroyed outside per-file cleanup'
profile = json.loads((base / 'destruction.json').read_text())
rows = []
for type_index, name in enumerate(labels):
    if not counts[type_index]:
        continue
    function = 'profile_free_' + name
    calls = sum(row['calls'] for row in profile['edges'] if row['callee'] == function)
    assert calls == counts[type_index], f'{name} counter/profile mismatch'
    row = next(row for row in profile['inclusive'] if row['function'] == function)
    rows.append(dict(type=name, destroyed=calls, instructions=row['inclusive'], per_object=row['inclusive']/calls,
                     during_collect=phases[2, type_index], file_cleanup=phases[3, type_index], after_files=phases[4, type_index], bytes_at_death=bytes_at_death[type_index]))
result = dict(allocated=allocated, destroyed=sum(counts.values()), peak_nodes=nodes['peak'],
              release_states={['null', 'immortal', 'nonfinal', 'final'][state]: sum(row['release'] for key, row in references.items() if key[1] == state) for state in range(4)},
              by_type=rows, references=[references[key] for key in sorted(references)])
(base / 'destruction-summary.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({key: value for key, value in result.items() if key != 'references'}, indent=2))
