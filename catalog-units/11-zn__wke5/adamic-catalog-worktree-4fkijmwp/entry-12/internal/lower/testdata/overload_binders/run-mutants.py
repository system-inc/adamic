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
    ('additional-binder-refusal', 'declaredTypes, servedTypes := overload.TypeParameters(), implementation.TypeParameters()', 'declaredTypes, servedTypes := overload.TypeParameters(), implementation.TypeParameters(); if len(servedTypes) > len(declaredTypes) { return l.notYet(overload, label + \" with additional implementation type parameters\") }', 'TestNativeAgreesWithNode/internal/oracle/testdata/(census_overload_binders|overload_(array_to_|ancestor_directory|leading_comment_range|trailing_comment_range|original_node|mutate_map|resolve_type_names|sort_deduplicate))'),
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
        if name in ('parameter', 'result'):
            names = ['serializer'] if name == 'parameter' else ['token', 'trampoline']
            selection = 'TestOverloadContractRulings/(' + '|'.join(names) + ')'
            with (logs / f'{name}-contracts.log').open('w') as log:
                result = subprocess.run(['go', 'test', './internal/oracle', '-run', selection, '-count=1', '-timeout', '10m'], cwd=root, stdout=log, stderr=subprocess.STDOUT, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'})
            output = (logs / f'{name}-contracts.log').read_text()
            if result.returncode == 0 or '[build failed]' in output or any('--- FAIL: TestOverloadContractRulings/' + fixture not in output for fixture in names):
                raise SystemExit(f'{name}: a refusal fixture survived: {output}')
            print(f'{name}: also caught by {len(names)} Node-held refusal fixtures', flush=True)
        if name == 'additional-binder-refusal':
            fixtures = ['array_to_map', 'array_to_multimap', 'array_to_numeric_map', 'ancestor_directory', 'leading_comment_range', 'trailing_comment_range', 'original_node', 'mutate_map', 'mutate_map_skipping_new', 'resolve_type_names', 'sort_deduplicate']
            if any('overload_' + fixture + '.a' not in output for fixture in fixtures):
                raise SystemExit('additional-binder-refusal: a per-kind fixture survived')
            print('additional-binder-refusal: all 11 executable per-kind fixtures caught it', flush=True)

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
        if name in ('union-inference', 'optional-rigid-binder'):
            fixtures = ['array_to_map', 'array_to_multimap', 'array_to_numeric_map'] if name == 'union-inference' else ['ancestor_directory']
            selection = 'TestNativeAgreesWithNode/internal/oracle/testdata/overload_(' + '|'.join(fixtures) + ')'
            with (logs / f'{name}-oracle.log').open('w') as log:
                result = subprocess.run(['go', 'test', './internal/oracle', '-run', selection, '-count=1', '-timeout', '10m'], cwd=root, stdout=log, stderr=subprocess.STDOUT, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'})
            output = (logs / f'{name}-oracle.log').read_text()
            if result.returncode == 0 or '[build failed]' in output or any('overload_' + fixture + '.a' not in output for fixture in fixtures):
                raise SystemExit(f'{name}: a fixture survived: {output}')
            print(f'{name}: also caught by {len(fixtures)} oracle fixtures', flush=True)

    finally:
        source.write_text(original)
