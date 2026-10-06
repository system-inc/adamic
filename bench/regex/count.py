#!/usr/bin/env python3
"""Count native instructions over one complete input cycle after untimed warmup."""
import json
import os
import pathlib
import subprocess
import sys

root = pathlib.Path(sys.argv[1])
cases = json.loads((root / 'cases.json').read_text())
valgrind = os.environ.get('CALLGRIND', 'valgrind')
print('pattern,instructions,calls', flush=True)
for index, case in enumerate(cases):
    destination = root / f'callgrind.{index}'
    with (root / f'callgrind.{index}.log').open('w') as log:
        subprocess.run([valgrind, '--tool=callgrind', '--instr-atstart=no',
                        f'--callgrind-out-file={destination}', str(root / 'native'),
                        str(index), str(len(case['Inputs']))], stdout=log, stderr=log,
                       check=True, timeout=120)
    counts = []
    for dump in [destination, *root.glob(f'callgrind.{index}.*')]:
        if dump.suffix == '.log':
            continue
        for line in dump.read_text().splitlines():
            # Explicit client dumps report zero in summary but the correct totals.
            if line.startswith('totals:'):
                counts.append(int(line.split()[1]))
    if not counts or max(counts) == 0:
        raise SystemExit(f'missing instruction counts: {case["Name"]}')
    print(f'{case["Name"]},{max(counts)},{len(case["Inputs"])}', flush=True)
