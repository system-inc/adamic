#!/usr/bin/env python3
"""Run four independent Uint16 conversion mutants and restore every edited file."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
logs = Path('/tmp/new-expression-uint16-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('constructor-conversion', 'internal/native/runtime/uint16_array.c',
     'adamic_bitwise_and(numbers->elements[index].number, 65535.0)', 'numbers->elements[index].number'),
    ('write-conversion', 'internal/native/emit_statements.go',
     'value = uint16Value(value)', 'value = value'),
    ('fill-conversion', 'internal/native/typed_arrays.go',
     'value = uint16Value(value)', 'value = value'),
    ('javascript-kind', 'internal/javascript/typed_arrays.go',
     'case ir.Uint16Array:\n\t\treturn "Uint16Array"', 'case ir.Uint16Array:\n\t\treturn "Uint8Array"'),
]
command = ['go', 'test', './internal/oracle', '-run',
           'TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_uint16',
           '-count=1', '-timeout', '10m']
env = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
for name, file, before, after in mutants:
    path = root / file
    original = path.read_bytes()
    source = original.decode()
    assert source.count(before) == 1, (name, 'mutant anchor changed')
    try:
        path.write_text(source.replace(before, after))
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run(command, cwd=root, env=env, stdout=output, stderr=subprocess.STDOUT)
        log = (logs / (name + '.log')).read_text()
        assert result.returncode != 0, (name, 'survived')
        assert 'stdout differs' in log, (name, 'not killed by output comparison', log)
        assert 'clang failed' not in log, (name, 'invalid build kill', log)
        print(name + ': killed by Node stdout comparison', flush=True)
    finally:
        path.write_bytes(original)
