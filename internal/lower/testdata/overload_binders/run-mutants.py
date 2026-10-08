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
