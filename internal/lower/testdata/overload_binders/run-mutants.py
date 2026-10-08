#!/usr/bin/env python3
"""Run each overload witness mutant independently and restore the compiler."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[4]
source = root / 'internal/lower/census_small.go'
original = source.read_text()
start = original.index('func (l *lowering) censusOverload(')
end = original.index('\nfunc (l *lowering) censusOverloadParameter', start)
body = original[start:end]
logs = Path('/tmp/notyet-overloads-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('parameter-inference', 'l.inferTypes(l.checker.GetTypeAtLocation(parameter), l.checker.GetTypeAtLocation(overload.Parameters()[index]), inferred)', '_ = parameter // missing parameter inference', 'TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders'),
    ('bottom-witness', 'l.checker.GetNonNullableType(l.checker.GetUndefinedType())', 'l.checker.GetTypeAtLocation(implementation.Name())', 'TestNativeAgreesWithNode/internal/oracle/testdata/census_overload_binders'),
    ('constraint', 'constraint != nil && !l.censusRelated(targets[index], constraint)', 'false && constraint != nil && !l.censusRelated(targets[index], constraint)', 'TestCensusOverloadBinderGuards/constraint'),
    ('parameter', 'if !l.censusRelated(given, takes) {', 'if false && !l.censusRelated(given, takes) {', 'TestCensusOverloadBinderGuards/parameter'),
    ('result', 'if !l.censusRelated(produced, promised) {', 'if false && !l.censusRelated(produced, promised) {', 'TestCensusOverloadBinderGuards/result'),
]
for name, before, after, test in mutants:
    if body.count(before) != 1:
        raise SystemExit(f'{name}: mutation anchor is not unique')
    package = './internal/oracle' if test.startswith('TestNative') else './internal/lower'
    try:
        source.write_text(original[:start] + body.replace(before, after) + original[end:])
        with (logs / f'{name}.log').open('w') as log:
            result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m'], cwd=root, stdout=log, stderr=subprocess.STDOUT, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'})
        output = (logs / f'{name}.log').read_text()
        if result.returncode == 0 or '--- FAIL:' not in output or '[build failed]' in output:
            raise SystemExit(f'{name}: survived or failed outside the intended test: {output}')
        print(f'{name}: caught by {test}; exit {result.returncode}', flush=True)
    finally:
        source.write_text(original)

source = root / 'internal/lower/generic.go'
original = source.read_text()
start = original.index('func (l *lowering) inferTypes(')
end = original.index('\n// JSON descriptors', start)
body = original[start:end]
mutants = [
    ('union-inference', 'if missing != nil {\n\t\t\tinto[missing] = l.concrete(instantiated)', 'if false && missing != nil {\n\t\t\tinto[missing] = l.concrete(instantiated)', 'one_compatible_binder'),
    ('known-union-member', 'if !l.checker.IsTypeAssignableTo(known, l.concrete(instantiated)) {', 'if false && !l.checker.IsTypeAssignableTo(known, l.concrete(instantiated)) {', 'incompatible_known_member'),
    ('multiple-union-binders', 'return // More than one unknown binder has no unique witness.', '// accept the last unknown binder', 'multiple_unknown_binders'),
    ('optional-rigid-binder', 'l.inferTypes(present[0], given, into)', 'l.inferTypes(l.checker.GetNonNullableType(declared), l.checker.GetNonNullableType(instantiated), into)', 'optional_rigid_binder'),
]
for name, before, after, subtest in mutants:
    if body.count(before) != 1:
        raise SystemExit(f'{name}: mutation anchor is not unique')
    try:
        source.write_text(original[:start] + body.replace(before, after) + original[end:])
        with (logs / f'{name}.log').open('w') as log:
            result = subprocess.run(['go', 'test', './internal/lower', '-run', f'TestOverloadInferenceWitnesses/{subtest}', '-count=1', '-timeout', '10m'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        output = (logs / f'{name}.log').read_text()
        if result.returncode == 0 or '--- FAIL:' not in output or '[build failed]' in output:
            raise SystemExit(f'{name}: survived or failed outside the intended test: {output}')
        print(f'{name}: caught by TestOverloadInferenceWitnesses/{subtest}; exit {result.returncode}', flush=True)
    finally:
        source.write_text(original)
