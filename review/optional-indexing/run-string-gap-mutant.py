#!/usr/bin/env python3
"""Restore the former string-index refusal and require the positive pin to fail."""
import json
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
path = root / 'internal/lower/object.go'
original = path.read_text()
needle = '\tif optional && (object.Type() == ir.Array || object.Type() == ir.String || object.Type().IsTypedArray()) {'
assert original.count(needle) == 1
mutated = original.replace(needle, '\tif optional && object.Type() == ir.String {\n\t\treturn nil, l.notYet(node, "?.[] on a string")\n\t}\n' + needle)
command = ['go', 'test', './internal/oracle', '-run', '^TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string_gap.a$', '-count=1']
log_path = Path('/tmp/optional-indexing-string-gap-mutant.log')
try:
    path.write_text(mutated)
    with log_path.open('w') as log:
        result = subprocess.run(command, cwd=root, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'}, stdout=log, stderr=subprocess.STDOUT)
finally:
    path.write_text(original)
output = log_path.read_text()
catcher = "stage 0 can't lower ?.[] on a string yet"
assert result.returncode == 1 and '--- FAIL:' in output and catcher in output and '[build failed]' not in output, output
(root / 'review/optional-indexing/string-gap-mutant.json').write_text(json.dumps({'mutation': 'restore former optional string indexing refusal', 'command': command, 'exit': result.returncode, 'catcher': catcher}, indent=2) + '\n')
print('Former string-index refusal: caught by positive Node fixture, exit 1')
