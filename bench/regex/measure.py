#!/usr/bin/env python3
"""Interleave two prebuilt native binaries and warmed Node 24, best of five."""
import argparse
import json
import math
import pathlib
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('before', type=pathlib.Path)
parser.add_argument('after', type=pathlib.Path)
parser.add_argument('--short-call-multiplier', type=int, default=1,
                    help='multiply rows with at least 9000 calls to reduce short-loop noise')
parser.add_argument('--minimum-native-seconds', type=float, default=0,
                    help='use an untimed baseline pilot to size short rows for this duration')
args = parser.parse_args()
if args.short_call_multiplier < 1:
    parser.error('short-call-multiplier must be positive')
if args.minimum_native_seconds < 0 or not math.isfinite(args.minimum_native_seconds):
    parser.error('minimum-native-seconds must be finite and nonnegative')
before, after = args.before, args.after
cases = json.loads((after / 'cases.json').read_text())
if cases != json.loads((before / 'cases.json').read_text()):
    raise SystemExit('benchmark cases differ')
print('pattern,native_before_seconds,native_after_seconds,node_seconds,calls,checksum', flush=True)
for index, case in enumerate(cases):
    calls = case['Calls'] * (args.short_call_multiplier if case['Calls'] >= 9000 else 1)
    if args.minimum_native_seconds and case['Calls'] >= 9000:
        pilot = subprocess.run([str(before / 'native'), str(index), str(case['Calls'])],
                               capture_output=True, text=True, check=True, timeout=120)
        seconds = float(pilot.stdout.split()[0])
        if seconds <= 0:
            raise SystemExit(f'missing pilot duration: {case["Name"]}')
        calls = max(calls, case['Calls'] * math.ceil(args.minimum_native_seconds / seconds))
    commands = [
        [str(before / 'native'), str(index), str(calls)],
        [str(after / 'native'), str(index), str(calls)],
        ['node', str(after / 'node.cjs'), str(index), str(calls)],
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
          f'{calls},{checksum}', flush=True)
