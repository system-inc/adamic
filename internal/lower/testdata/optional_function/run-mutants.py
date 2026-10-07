#!/usr/bin/env python3
"""Mutate one optional callable rule at a time, with source restored in finally."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
logs = Path('/tmp/adamic-optional-function-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('present-undefined-false', 'internal/lower/control.go',
     'return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: condition}}, nil',
     'return ir.IsUndefined{Value: condition}, nil',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/optional_function_presence', 'stdout differs'),
    ('present-null-false', 'internal/lower/control.go',
     'return ir.Unary{Operator: ir.Not, Operand: ir.IsNull{Value: condition}}, nil',
     'return ir.IsNull{Value: condition}, nil',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/optional_function_presence', 'stdout differs'),
    ('required-allowed', 'internal/lower/control.go',
     '(condition.Type() == ir.Object || condition.Type() == ir.Closure) && l.includesUndefined(proven)',
     '(condition.Type() == ir.Object && l.includesUndefined(proven)) || condition.Type() == ir.Closure',
     './internal/lower', 'TestOptionalFunctionKeepsRequiredRefusal', 'got <nil>'),
    ('nullable-refused', 'internal/lower/expression.go',
     'if l.nullableCallable(proven) {',
     'if false {',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/optional_function_presence', 'Lower:'),
    ('nullable-object-allowed', 'internal/lower/optional_function.go',
     'if !known || held != ir.Closure {',
     'if !known || held == ir.Number {',
     './internal/lower', 'TestOptionalFunctionNullableObjectIsNotYet', 'got <nil>'),
    ('mixed-absence-allowed', 'internal/lower/expression.go',
     'if l.includesUndefined(proven) {\n\t\t\t\treturn 0, false\n\t\t\t}',
     '',
     './internal/lower', 'TestOptionalFunctionMixedAbsenceIsNotYet', 'got <nil>'),
]
for name, file, old, new, package, test, catcher in mutants:
    path = root / file
    original = path.read_text()
    assert original.count(old) == 1, name
    try:
        path.write_text(original.replace(old, new))
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', '-count=1', '-timeout', '10m', package, '-run', test], cwd=root, stdout=log, stderr=subprocess.STDOUT, env=os.environ | {'ADAMIC_GATE_UNCACHED': '1'})
        output = (logs / (name + '.log')).read_text()
        assert result.returncode != 0 and catcher in output, f'{name} not caught: {output}'
        assert 'build failed' not in output and 'clang failed' not in output, output
        print(f'{name}: caught by {catcher}, exit {result.returncode}', flush=True)
    finally:
        path.write_text(original)
