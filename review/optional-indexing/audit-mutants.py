#!/usr/bin/env python3
"""Prove the preserved census audit rejects altered counts and compiler hashes."""
import json
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
results = []
for name, filename, mutate, catcher in [
    ('census-count', 'census.json', lambda value: value['rows'][0].__setitem__('after', value['rows'][0]['after'] + 1), 'recorded census changed'),
    ('compiler-hash', 'compiler-hashes.json', lambda value: value.__setitem__(next(iter(value)), '0' * 64), 'AssertionError'),
]:
    path = root / 'review/optional-indexing' / filename
    original = path.read_bytes()
    value = json.loads(original)
    mutate(value)
    try:
        path.write_text(json.dumps(value))
        command = [sys.executable, 'review/optional-indexing/summarize-census.py', '--audit']
        with Path('/tmp/optional-indexing-audit-' + name + '.log').open('w') as log:
            run = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT)
    finally:
        path.write_bytes(original)
    output = Path('/tmp/optional-indexing-audit-' + name + '.log').read_text()
    assert run.returncode == 1 and catcher in output, output
    results.append({'name': name, 'exit': run.returncode, 'catcher': catcher})
    print(name + ': caught by ' + catcher)
(root / 'review/optional-indexing/audit-mutants.json').write_text(json.dumps(results, indent=2) + '\n')
