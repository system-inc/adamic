#!/usr/bin/env python3
"""Run independent never lowering mutants, restoring the emitter each time."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[3]
lowering = root / 'internal/lower/expression.go'
backend = root / 'internal/javascript/taste.go'
logs = Path('/tmp/adamic-never-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('drop-fit', lowering, 'effects.Result = ir.Unreachable{Of: to}',
     'effects.Result = ir.Unreachable{}', './internal/oracle',
     'TestNativeAgreesWithNode/internal/oracle/testdata/never_string',
     '-Wincompatible-pointer-types'),
    ('drop-fit-ir', lowering, 'effects.Result = ir.Unreachable{Of: to}',
     'effects.Result = ir.Unreachable{}', './internal/lower',
     'TestNeverFitsWithoutConversion', 'want destination marker'),
    ('drop-call-js', backend, 'e.statements(expression.Body)',
     'if _, never := expression.Result.(ir.Unreachable); !never { e.statements(expression.Body) }',
     './internal/oracle',
     'TestNativeAgreesWithNode/internal/oracle/testdata/never_values',
     'stdout differs'),
]
originals = {path: path.read_text() for path in (lowering, backend)}
try:
    for name, source, before, after, package, test, catcher in mutants:
        original = originals[source]
        assert original.count(before) == 1, name
        source.write_text(original.replace(before, after, 1))
        path = logs / (name + '.log')
        with path.open('w') as log:
            result = subprocess.run(
                ['go', 'test', package, '-run', test,
                 '-count=1', '-timeout', '10m'], cwd=root,
                env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'},
                stdout=log, stderr=subprocess.STDOUT)
        output = path.read_text()
        assert result.returncode != 0, name + ' survived'
        assert catcher in output, name + ' missed intended catcher: ' + str(path)
        print(name + ': caught by ' + catcher + ', exit ' + str(result.returncode), flush=True)
        source.write_text(original)
finally:
    for source, original in originals.items():
        source.write_text(original)
