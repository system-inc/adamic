#!/usr/bin/env python3
"""Isolate private iterator-view rules and restore every mutant."""
from pathlib import Path
import os
import subprocess
import sys

repository = Path(__file__).resolve().parents[3]
source = repository / 'internal/lower/for_of_library_view.go'
original = source.read_text()
logs = Path('/tmp/adamic-for-of-library-view-mutants')
logs.mkdir(exist_ok=True)
guards = 'TestForOfLibraryViewOriginChecks'
oracle = 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_library_view'
escaped = '''if parent == nil || parent.Kind != ast.KindCallExpression || parent.AsCallExpression().Expression != node {
					sound = false
					return true
				}'''
assert original.count(escaped) == 1
mutants = [
    ('omit-view-admission', original.replace('return element, sound && called', 'return element, sound && called && false'), './internal/oracle', oracle, 'for...of over an object'),
    ('unchecked-producer', original.replace('if !l.forOfLibraryIteratorFactory(argument, element) {', 'if !l.forOfLibraryIteratorFactory(argument, element) && false {'), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('ignore-function-escape', original.replace(escaped, escaped.replace('\n\t\t\t\t\tsound = false', '').replace('return true', 'return false')), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('ignore-parameter-use', original.replace('if referenced == symbol && node != parameter.Name() {', 'if referenced == symbol && node != parameter.Name() && false {'), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('allow-export', original.replace(' || ast.HasSyntacticModifier(owner, ast.ModifierFlagsExport)', ''), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('allow-default', original.replace(' || declared.Initializer != nil', ''), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('allow-generic', original.replace(' || len(owner.TypeParameters()) != 0', ''), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('allow-maybe-element', original.replace(' || (element != ir.String && element != ir.Number && element != ir.Boolean && element != ir.Object)', ''), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('allow-no-callers', original.replace('return element, sound && called', 'return element, sound || called'), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('ignore-spread-position', original.replace('if argument.Kind == ast.KindSpreadElement {', 'if argument.Kind == ast.KindSpreadElement && false {'), './internal/lower', guards, 'want the unproved iterator-view stop'),
    ('read-type-as-storage', original.replace('l.kept(arguments[0])', 'l.representation(arguments[0])'), './internal/lower', guards, 'want the unproved iterator-view stop'),
]
mutants = [(name, source, changed, package, test, catcher) for name, changed, package, test, catcher in mutants]
library = repository / 'internal/lower/library_map_set.go'
library_original = library.read_text()
binding = 'bindings = append(bindings, ir.Declare{Local: local, Value: item})'
assert library_original.count(binding) == 1
shared = library_original.replace('var bindings []ir.Statement', 'var bindings, shared []ir.Statement', 1)
shared = shared.replace(binding, '''var initial ir.Expression
		switch l.result.Locals[local].Type {
		case ir.Object:
			initial = ir.ObjectLiteral{}
		case ir.String:
			initial = ir.StringConstant{Index: l.constant("")}
		case ir.Number:
			initial = ir.NumberConstant{}
		default:
			initial = ir.BooleanConstant{}
		}
		shared = append(shared, ir.Declare{Local: local, Value: initial})
		bindings = append(bindings, ir.Assign{Local: local, Value: item})''', 1)
shared = shared.replace('Body: l.libraryIteratorLoop(l.functionIndex, iterator, item, append(bindings, body...))', 'Body: append(shared, l.libraryIteratorLoop(l.functionIndex, iterator, item, append(bindings, body...))...)', 1)
mutants += [
    ('shared-iterator-binding', library, shared, './internal/oracle', oracle, 'stdout differs'),
    ('premature-exhaustion', library, library_original.replace('Condition: ir.Property{Object: step, Name: "done", Of: ir.Boolean}', 'Condition: ir.BooleanConstant{Value: true}', 1), './internal/oracle', oracle, 'stdout differs'),
]
for name, path, changed, package, test, catcher in mutants:
    if len(sys.argv) > 1 and name not in sys.argv[1:]:
        continue
    before = path.read_text()
    assert changed != before, name
    try:
        path.write_text(changed)
        log = logs / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m'], cwd=repository, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and catcher in observed, (name, observed)
        assert '[build failed]' not in observed and '[-Werror' not in observed and 'panic:' not in observed, (name, observed)
        print(f'{name}: caught ({catcher}), exit {result.returncode}', flush=True)
    finally:
        path.write_text(before)
