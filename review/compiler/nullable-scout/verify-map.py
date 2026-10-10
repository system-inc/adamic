#!/usr/bin/env python3
"""Compare the map's reason counts with the pinned census, independently."""
import json
import pathlib
import subprocess
import sys
root = pathlib.Path(__file__).parent
source = json.loads(subprocess.check_output(['git', 'show', '807d65d9:stage3/census/tsc-closure/rerun-2026-10-09/RESULT.json']))
mapped = json.loads(pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else root / 'census-nullable-reasons.json').read_text())
rows = [row for group in mapped['groups'].values() for row in group] + mapped['excluded']
assert len({row['reason'] for row in rows}) == len(rows), 'duplicate reason'
for mode in ['full', 'latent']:
    reasons = source[mode]['all']['per_reason']
    selected = {reason: count for reason, count in reasons.items() if any(word in reason for word in ['undefined', 'null', 'optional field', 'nullable'])}
    observed = {row['reason']: row[mode] for row in rows if row[mode]}
    assert observed == selected, f'{mode}: missing, added or changed census reason/count'
print('all selected full and latent reasons/counts match pinned census')
