#!/usr/bin/env python3
"""Verify recorded stops and Node-running minimal witnesses."""
import argparse
import json
from pathlib import Path
import re
import sys

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('evidence', type=Path, nargs='?', default=Path(__file__).resolve().parent / 'evidence')
args = parser.parse_args()
evidence = args.evidence
records = json.loads((evidence / 'stops.json').read_text())
failures = []
if len(records) != 15:
    failures.append('stop population')
for record in records:
    ordinal = record['ordinal']
    stem = Path(record['probe']).stem
    stdout = (evidence / f'{stem}-node.stdout').read_text()
    stderr = (evidence / f'{stem}-node.stderr').read_text()
    node_exit = int((evidence / f'{stem}-node.exit').read_text())
    if stdout != record['node_stdout'] or stderr or node_exit != 0:
        failures.append(f'{ordinal}: Node observation')
    native_exit = int((evidence / f'{stem}-build.exit').read_text())
    native = (evidence / f'{stem}-build.stderr').read_text().splitlines()[0]
    actual = re.sub(r'^.+:\d+:\d+: ', '', native)
    if native_exit != 1 or actual != record['probe_message']:
        failures.append(f'{ordinal}: probe diagnostic')
    for split in (0, 1):
        actual = (evidence / f'{ordinal:02d}-split-{split}.stderr').read_text().splitlines()[0]
        actual = re.sub(r'^.+:\d+:\d+: ', '', actual)
        if actual != record['message'] or int((evidence / f'{ordinal:02d}-split-{split}.exit').read_text()) != 1:
            failures.append(f'{ordinal}: split {split} diagnostic')
    if (evidence / f'{ordinal:02d}-split-0.stderr').read_bytes() != (evidence / f'{ordinal:02d}-split-1.stderr').read_bytes():
        failures.append(f'{ordinal}: split byte comparison')
if failures:
    print('\n'.join(failures), file=sys.stderr)
    sys.exit(1)
print('15 stops verified; 15 Node witnesses exit 0; both split modes agree')
