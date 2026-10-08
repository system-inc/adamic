#!/usr/bin/env python3
"""Run a component implementation and compare stdout, empty stderr and exit zero independently."""
import argparse
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--output', required=True, type=Path)
parser.add_argument('--timeout', type=float, default=120)
parser.add_argument('golden', type=Path)
parser.add_argument('command', nargs=argparse.REMAINDER)
args = parser.parse_args()
command = args.command[1:] if args.command[:1] == ['--'] else args.command
if not command:
    parser.error('command required')
if args.output.exists():
    parser.error('output must be new')
wanted = args.golden.read_bytes()
args.output.mkdir(parents=True)
(args.output / 'command.json').write_text(json.dumps(command) + '\n')
with (args.output / 'actual.stdout').open('wb') as out, (args.output / 'actual.stderr').open('wb') as err:
    try:
        result = subprocess.run(command, stdout=out, stderr=err, timeout=args.timeout)
        code = result.returncode
        timeout = False
    except subprocess.TimeoutExpired:
        code, timeout = 124, True
(args.output / 'actual.exit').write_text(str(code) + '\n')
expected = {'stdout': wanted, 'stderr': b'', 'exit': b'0\n'}
differences = {}
for stream, value in expected.items():
    actual = (args.output / ('actual.' + stream)).read_bytes()
    if value != actual:
        offset = next((i for i, (a, b) in enumerate(zip(value, actual)) if a != b), min(len(value), len(actual)))
        differences[stream] = {'first_byte': offset, 'expected_bytes': len(value), 'actual_bytes': len(actual)}
report = {'success': not differences and not timeout, 'timeout': timeout, 'differences': differences}
(args.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(report))
raise SystemExit(0 if report['success'] else 1)
