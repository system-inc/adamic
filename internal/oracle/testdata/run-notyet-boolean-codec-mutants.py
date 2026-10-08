#!/usr/bin/env python3
"""Prove optional boolean slot presence, value and evaluation rules against the Node fixture."""
from pathlib import Path
import json, os, subprocess
root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/destructuring-morning-codec-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('optional-undefined-pack', 'internal/native/destructuring_boolean_slots.go', 'return !value.present ? 2.0', 'return !value.present ? 0.0'),
    ('optional-value-unpack', 'internal/native/destructuring_boolean_slots.go', 'packed == 1.0', 'packed == 0.0'),
    ('optional-pack-once', 'internal/native/emit_objects.go', 'values := make([]string, 0, len(literal.Fields))\n\tfor _, field := range literal.Fields {\n\t\tif region {', 'values := make([]string, 0, len(literal.Fields))\n\tfor _, field := range literal.Fields {\n if field.Value.Type() == ir.MaybeBoolean { e.value(field.Value) }\n\t\tif region {'),
]
for name, path, before, after in mutants:
    file = root / path
    original = file.read_text()
    assert original.count(before) == 1
    source = logs / (name + '.go')
    source.write_text(original.replace(before, after, 1))
    overlay = logs / (name + '.json')
    overlay.write_text(json.dumps({'Replace': {str(file): str(source)}}))
    with (logs / (name + '.log')).open('w') as out:
        result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/oracle', '-run', '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_destructuring_option_flags.a$', '-count=1', '-timeout', '10m'], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=out, stderr=subprocess.STDOUT)
    output = (logs / (name + '.log')).read_text()
    assert result.returncode != 0 and 'stdout differs' in output and '[build failed]' not in output, (name, result.returncode, output)
    print(name + ': caught by stdout differs', flush=True)
