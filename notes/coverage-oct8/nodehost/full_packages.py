#!/usr/bin/env python3
"""For surviving focused mutants, run complete native/lower packages.
Run after mutants.py has finished. Each mutation is restored in finally.
"""
import json
import os
from pathlib import Path
import subprocess

logs = Path('/tmp/nodehost-mutants')
cases = json.loads((logs / 'summary.json').read_text())
results = {}
for name, case in cases.items():
    if case['packages_exit'] != 0:
        continue
    path = Path(case['file'])
    original = path.read_text()
    start = original.index('static adamic_array *read_buffer(') if name == 'owned-buffer-close' else original.index('static double write_data(') if name == 'short-write-loop' else 0
    at = original.index(case['removed'], start)
    try:
        path.write_text(original[:at] + case['replacement'] + original[at + len(case['removed']):])
        with (logs / (name + '.full-packages.log')).open('wb') as log:
            result = subprocess.run(['go', 'test', './internal/lower', './internal/native', '-count=1', '-timeout', '30m'], env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=log, stderr=subprocess.STDOUT)
        results[name] = result.returncode
        (logs / 'full-packages.json').write_text(json.dumps(results, indent=2) + '\n')
        print(name, 'full native/lower packages', result.returncode, flush=True)
    finally:
        path.write_text(original)
