#!/usr/bin/env python3
"""Falsify each pair's expected arity without editing compiler/runtime sources."""
import json
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[4]
logs = Path('/tmp/lane5-b-mutants') / Path(os.environ.get('ADAMIC_CALLABLE_SHARE_B_FAMILIES', 'families.json')).stem
logs.mkdir(parents=True, exist_ok=True)
families = json.loads((Path(__file__).parent / os.environ.get('ADAMIC_CALLABLE_SHARE_B_FAMILIES', 'families.json')).read_text())
controls = logs / 'controls.log'
with controls.open('w') as log:
    result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewCallableShareBArityMutants$', '-count=1', '-v', '-timeout', '5m'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
assert result.returncode == 0, controls.read_text()
mutants = logs / 'arity.log'
environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1', ADAMIC_CALLABLE_SHARE_B_MUTANT='arity')
with mutants.open('w') as log:
    result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewCallableShareBFamilies$/^rank-/^wrong-arity$', '-count=1', '-v', '-timeout', '5m'], cwd=root, env=environment, stdout=log, stderr=subprocess.STDOUT)
text = mutants.read_text()
assert result.returncode != 0, text
assert '[build failed]' not in text and 'clang failed:' not in text, text
for family in families:
    assert '--- FAIL: TestCheckedViewCallableShareBFamilies/' + family['directory'] + '/wrong-arity' in text, family['rank']
print(str(len(families)) + ' arity mutants caught by the original exit/message pins; both-backend valid-execution controls pass.')
