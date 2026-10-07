#!/usr/bin/env python3
"""Prove CLI comparison independently catches a column, stderr, and exit mutant."""
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import sys
import tempfile
from corpus import ROOT
from driver import command_argv

command = command_argv(sys.argv[1:])
results = Path(tempfile.mkdtemp(prefix='tsc-mutants-'))
print('results: ' + str(results), flush=True)
for suffix in ('stdout', 'stderr', 'exit'):
    golden = results / suffix / 'goldens' / 'tiny'
    golden.mkdir(parents=True)
    for field in ('stdout', 'stderr', 'exit'):
        shutil.copyfile(ROOT / 'tiny' / ('golden.' + field), golden / ('golden.' + field))
    target = golden / ('golden.' + suffix)
    before = target.read_bytes()
    if suffix == 'stdout':
        after, changed = re.subn(rb'\((\d+),(\d+)\)', lambda m: b'(' + m[1] + b',' + str(int(m[2]) + 1).encode() + b')', before, count=1)
        assert changed == 1
    elif suffix == 'stderr':
        after = b'changed stderr\n'
    else:
        after = b'0\n'
    assert after != before
    target.write_bytes(after)
    env = dict(os.environ, TSC_RESULTS=str(results / suffix / 'run'))
    env.pop('NATIVE_TSC', None)
    with (results / (suffix + '.log')).open('wb') as log:
        completed = subprocess.run([str(ROOT / 'run.sh'), '--tiny', '--golden-root', str(golden.parent), '--', *command], env=env, stdout=log, stderr=log)
    report = json.loads((results / suffix / 'run/report.json').read_text())
    assert completed.returncode == 1 and report['failed'] == ['tiny'], suffix + ' mutant survived'
    log = (results / (suffix + '.log')).read_text()
    assert 'FAIL tiny: ' + suffix in log
    print(f'CAUGHT golden {suffix} mutant by exact {suffix} comparison; harness exit=1', flush=True)

# The native slot is exercised by a transparent Node forwarding executable.
# This verifies plumbing, not a native compiler implementation.
forwarder = results / 'node-forwarder'
forwarder.write_text('#!/usr/bin/env bash\nexec ' + shlex.join(command) + ' "$@"\n')
forwarder.chmod(0o755)
env = dict(os.environ, NATIVE_TSC=str(forwarder), TSC_RESULTS=str(results / 'slot'))
with (results / 'slot.log').open('wb') as log:
    completed = subprocess.run([str(ROOT / 'run.sh'), '--tiny', '--', *command], env=env, stdout=log, stderr=log)
assert completed.returncode == 0
assert json.loads((results / 'slot/native/report.json').read_text())['passed'] == 1
print('PASS native-slot plumbing using a Node forwarder; no native claim', flush=True)

# Independent evidence mutations, each restored before the next assertion.
from audit import audit
artifact = results / 'provenance'
shutil.copytree(ROOT, artifact, ignore=shutil.ignore_patterns('__pycache__'))
selection_path = artifact / 'selection.json'
selection = json.loads(selection_path.read_text())
first = artifact / 'corpus' / selection['cases'][0]['id']
error_row = next(row for row in selection['cases'] if row['baseline'])
error_case = artifact / 'corpus' / error_row['id']
probes = [
    ('source hash', first / 'input.a', lambda b: b + b'\n'),
    ('baseline hash', error_case / 'reference.errors.txt', lambda b: b + b'\n'),
    ('baseline summary', error_case / 'expected.stdout', lambda b: b.replace(b'error TS', b'error XX', 1)),
    ('expected stderr', first / 'expected.stderr', lambda b: b'wrong\n'),
    ('expected exit', first / 'expected.exit', lambda b: b'2\n'),
    ('case population', selection_path, lambda b: json.dumps({**selection, 'cases': selection['cases'][:-1]}).encode()),
    ('header options', selection_path, lambda b: json.dumps({**selection, 'cases': [{**selection['cases'][0], 'options': {'strict': 'changed'}}, *selection['cases'][1:]]}).encode()),
    ('baseline codes', selection_path, lambda b: json.dumps({**selection, 'cases': [{**row, 'codes': row['codes'] + [99999]} if row['id'] == error_row['id'] else row for row in selection['cases']]}).encode()),
]
for check, path, mutate in probes:
    before = path.read_bytes()
    path.write_bytes(mutate(before))
    try:
        audit(artifact)
    except AssertionError as error:
        assert check in str(error), (check, str(error))
        print('CAUGHT provenance mutant by ' + check, flush=True)
    else:
        raise AssertionError(check + ' mutant survived')
    finally:
        path.write_bytes(before)
