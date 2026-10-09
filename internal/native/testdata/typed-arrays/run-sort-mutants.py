"""Run each sort mutation against the Node-held C and compiler fixtures."""
from pathlib import Path
import os
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
path = root / 'internal/native/runtime/typed_array_sort.c'
original = path.read_text()
logs = Path('/tmp/adamic-typed-sort-mutants')
logs.mkdir(exist_ok=True)
string_order = '''static bool before_string(double left, double right) {
	char a[ADAMIC_NUMBER_FORMAT_MAX + 1], b[ADAMIC_NUMBER_FORMAT_MAX + 1];
	a[adamic_number_format(left, a)] = 0;
	b[adamic_number_format(right, b)] = 0;
	return strcmp(a, b) < 0;
}
'''
mutants = {
    'zeros-equal': original.replace('return signbit(left) && !signbit(right);', 'return false;'),
    'nan-first': original.replace('if (isnan(left)) { return false; }', 'if (isnan(left)) { return !isnan(right); }').replace('if (isnan(right)) { return true; }', 'if (isnan(right)) { return false; }'),
    'whole-parent': original.replace('adamic_typed_array *range = array;', 'adamic_typed_array *range = array->owner != NULL ? array->owner : array;'),
    'string-order': original.replace('static bool before_float64', string_order + '\nstatic bool before_float64').replace('return left < right;', 'return before_string(left, right);').replace('#define NUMERIC_BEFORE(left, right) ((left) < (right))', '#define NUMERIC_BEFORE(left, right) before_string(left, right)'),
}
if "--comparator-only" in sys.argv:
    mutants = {}
checks = [
    ('C', ['./internal/native', '-run', '^TestTypedArrayRuntime$']),
    ('backends', ['./internal/oracle', '-run', '^TestNativeAgreesWithNode$/internal/oracle/testdata/typed_arrays_sort']),
]
environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
try:
    for name, changed in mutants.items():
        assert changed != original, (name, 'stale mutant')
        path.write_text(changed)
        for check, arguments in checks:
            log = logs / f'{name}-{check}.log'
            with log.open('w') as output:
                result = subprocess.run(['go', 'test', *arguments, '-count=1', '-timeout', '10m'], cwd=root, env=environment, stdout=output, stderr=subprocess.STDOUT)
            report = log.read_text()
            expected = 'runtime differs from Node' if check == 'C' else 'stdout differs'
            if result.returncode == 0 or expected not in report or 'clang failed' in report or '[build failed]' in report:
                raise RuntimeError(f'{name} {check}: survived or wrong failure; see {log}')
            print(f'{name} {check}: caught by Node stdout difference; see {log}', flush=True)
        path.write_text(original)
finally:
    path.write_text(original)
if mutants:
    print('all 4 sort mutants caught by C and backend oracles')

# A bypass must be accepted incorrectly, so only the named NotYet assertion catches it.
lowering = root / 'internal/lower/typed_arrays.go'
original_lowering = lowering.read_text()
guard = '''		case "sort":
			if len(args) != 0 {
				return nil, true, l.notYet(node, "typed array sort with a comparator")
			}'''
assert original_lowering.count(guard) == 1
try:
    changed = original_lowering.replace(guard, '\t\tcase "sort":', 1)
    changed = changed.replace('\t\tlowered := []ir.Expression{}', '\t\tif method == "sort" { return ir.TypedArraySort{Array: array}, true, nil }\n\t\tlowered := []ir.Expression{}', 1)
    lowering.write_text(changed)
    log = logs / 'comparator-bypass-lower.log'
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/lower', '-run', '^TestTypedArraySortComparatorIsNotYet$', '-count=1', '-timeout', '10m'], cwd=root, env=environment, stdout=output, stderr=subprocess.STDOUT)
    report = log.read_text()
    if result.returncode == 0 or 'expected named comparator gap, got <nil>' not in report or '[build failed]' in report:
        raise RuntimeError(f'comparator bypass survived or failed incorrectly; see {log}')
    print(f'comparator bypass: caught by named NotYet assertion; see {log}')
finally:
    lowering.write_text(original_lowering)
