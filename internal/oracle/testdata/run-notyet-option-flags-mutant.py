#!/usr/bin/env python3
"""Destroy the optional boolean presence/value pair without changing production files."""
from pathlib import Path
import json, os, subprocess
root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/destructuring-morning-option-mutant')
logs.mkdir(exist_ok=True)
file = root / 'internal/lower/collections.go'
before = 'declaredStatements, err := l.declareDestructured(binding, fieldType, value)'
after = 'given := ir.Expression(value)\n if of == ir.MaybeBoolean { given = ir.MaybeOf{Of: ir.MaybeBoolean, Value: ir.BooleanConstant{Value: false}} }\n declaredStatements, err := l.declareDestructured(binding, fieldType, given)'
original = file.read_text()
assert original.count(before) == 1
source = logs / 'collections.go'
source.write_text(original.replace(before, after, 1))
overlay = logs / 'overlay.json'
overlay.write_text(json.dumps({'Replace': {str(file): str(source)}}))
with (logs / 'mutant.log').open('w') as out:
    result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/oracle', '-run', '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_destructuring_option_flags.a$', '-count=1', '-timeout', '10m'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=out, stderr=subprocess.STDOUT)
output = (logs / 'mutant.log').read_text()
assert result.returncode != 0 and 'stdout differs' in output and '[build failed]' not in output, (result.returncode, output)
print('optional-boolean-presence: caught by stdout differs')
