#!/usr/bin/env python3
"""Bypass the checked-storage view refusal without changing production files."""
from pathlib import Path
import json, os, subprocess
root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/destructuring-morning-view-mutant')
logs.mkdir(exist_ok=True)
file = root / 'internal/lower/destructuring_slot_views.go'
before = 'if of != ir.MaybeBoolean && of != ir.Union {'
after = 'if true || of != ir.MaybeBoolean && of != ir.Union {'
original = file.read_text()
assert original.count(before) == 1
source = logs / 'destructuring_slot_views.go'
source.write_text(original.replace(before, after, 1))
overlay = logs / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(file): str(source)}}))
with (logs / 'mutant.log').open('w') as out:
    result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/lower', '-run', '^TestDestructuringSlotViewsStayExplicit$', '-count=1', '-timeout', '10m'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=out, stderr=subprocess.STDOUT)
output = (logs / 'mutant.log').read_text()
assert result.returncode != 0 and 'want the checked-slot-view gap' in output and '[build failed]' not in output, (result.returncode, output)
for name in ['optional-boolean', 'boxed-union', 'tuple-assignment']:
    assert 'TestDestructuringSlotViewsStayExplicit/' + name in output, output
print('checked-storage-view: caught by want the checked-slot-view gap')
