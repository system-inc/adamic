#!/usr/bin/env python3
"""Interleave two prebuilt native binaries and warmed Node 24, best of five."""
import json
import pathlib
import subprocess
import sys

before, after = map(pathlib.Path, sys.argv[1:3])
cases = json.loads((after / 'cases.json').read_text())
if cases != json.loads((before / 'cases.json').read_text()):
    raise SystemExit('benchmark cases differ')
print('pattern,native_before_seconds,native_after_seconds,node_seconds,calls,checksum', flush=True)
for index, case in enumerate(cases):
    commands = [
        [str(before / 'native'), str(index), str(case['Calls'])],
        [str(after / 'native'), str(index), str(case['Calls'])],
        ['node', str(after / 'node.cjs'), str(index), str(case['Calls'])],
    ]
    best = [float('inf')] * 3
    checksum = None
    for round_number in range(5):
        # Rotate order to distribute machine drift across engines.
        for offset in range(3):
            engine = (round_number + offset) % 3
            result = subprocess.run(commands[engine], capture_output=True, text=True,
                                    check=True, timeout=120)
            seconds, observed = result.stdout.split()
            if checksum is not None and observed != checksum:
                raise SystemExit(f'DISAGREEMENT {case["Name"]}: {observed} != {checksum}')
            checksum = observed
            best[engine] = min(best[engine], float(seconds))
    print(f'{case["Name"]},{best[0]:.9f},{best[1]:.9f},{best[2]:.9f},'
          f'{case["Calls"]},{checksum}', flush=True)
