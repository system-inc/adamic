#!/usr/bin/env python3
"""Reject per-pattern instruction or wall-time regressions against a frozen VM."""
import argparse
import csv

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('before_instructions')
parser.add_argument('after_instructions')
parser.add_argument('--times')
args = parser.parse_args()

def read(path):
    with open(path) as stream:
        rows = list(csv.DictReader(stream))
    result = {row['pattern']: row for row in rows}
    if not result or len(result) != len(rows):
        raise SystemExit(f'missing or duplicated patterns: {path}')
    return result

before, after = read(args.before_instructions), read(args.after_instructions)
if before.keys() != after.keys():
    raise SystemExit('instruction pattern inventories differ')
failures = []
for name, row in before.items():
    old = int(row['instructions']) / int(row['calls'])
    new = int(after[name]['instructions']) / int(after[name]['calls'])
    if old <= 0 or new <= 0:
        failures.append(f'{name}: missing instruction count')
    elif new > old * 1.01:
        failures.append(f'{name}: INSTRUCTION REGRESSION {new / old:.4f}x')
if args.times:
    times = read(args.times)
    if times.keys() != before.keys():
        raise SystemExit('wall-time pattern inventories differ')
    for name, row in times.items():
        old, new = float(row['native_before_seconds']), float(row['native_after_seconds'])
        if old <= 0 or new <= 0:
            failures.append(f'{name}: missing wall time')
        elif new > old * 1.05:
            failures.append(f'{name}: WALL REGRESSION {new / old:.4f}x')
for failure in failures:
    print(failure)
if failures:
    raise SystemExit(1)
print(f'{len(before)} patterns: no instruction regressions above 1%' +
      (' or wall regressions above 5%' if args.times else ''))
