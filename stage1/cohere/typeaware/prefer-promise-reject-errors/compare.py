#!/usr/bin/env python3
"""Compare the isolated native suite with the unchanged production Go rules."""
import argparse
import json
import pathlib
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('oracle')
parser.add_argument('native')
parser.add_argument('cases')
parser.add_argument('--stop', action='store_true')
args = parser.parse_args()
failures = 0
findings = 0
for case in sorted(pathlib.Path(args.cases).glob('case-*')):
    if not (case / 'manifest').exists():
        continue
    outputs = []
    for name, binary in [('go', args.oracle), ('native', args.native)]:
        flags = json.loads((case / 'flags.json').read_text()) or []
        started = time.monotonic()
        completed = subprocess.run([binary, str(case / 'tsconfig.json'), str(case / 'manifest'), *flags], capture_output=True, timeout=90)
        (case / (name + '.stdout')).write_bytes(completed.stdout)
        (case / (name + '.stderr')).write_bytes(completed.stderr)
        outputs.append(completed)
    truth, native = outputs
    if truth.returncode or native.returncode or native.stderr or truth.stdout != native.stdout:
        failures += 1
        print('FAIL', case, (case / 'case.txt').read_text().splitlines()[0], 'exit', truth.returncode, native.returncode, flush=True)
        if native.stderr:
            print(native.stderr.decode(errors='replace')[:1200], flush=True)
        if truth.stdout != native.stdout:
            left = truth.stdout.splitlines()
            right = native.stdout.splitlines()
            for at in range(max(len(left), len(right))):
                expected = left[at] if at < len(left) else b''
                got = right[at] if at < len(right) else b''
                if expected != got:
                    print('Go:', expected.decode(errors='replace')[:1000], flush=True)
                    print('Native:', got.decode(errors='replace')[:1000], flush=True)
                    break
        if args.stop:
            break
    else:
        count = int(truth.stdout.splitlines()[-1].split()[-1])
        findings += count
        print('PASS', case.name, count, flush=True)
print('failures', failures, 'matching findings', findings, flush=True)
raise SystemExit(bool(failures))
