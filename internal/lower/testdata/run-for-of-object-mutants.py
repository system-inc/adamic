#!/usr/bin/env python3
"""Run isolated for-of object-binding mutants, restoring each source afterwards."""
from pathlib import Path
import os
import subprocess

repository = Path(__file__).resolve().parents[3]
logs = Path('/tmp/adamic-for-of-object-mutants')
logs.mkdir(exist_ok=True)
object_source = repository / 'internal/lower/object.go'
iteration_source = repository / 'internal/lower/iteration.go'
original_object = object_source.read_text()
original_iteration = iteration_source.read_text()
accessor_source = repository / 'internal/lower/class_accessors.go'
original_accessor = accessor_source.read_text()
needle = '''		if lowered.Body, err = l.destructureFrom(name, proven, element, lowered.Local); err != nil {
			return nil, err
		}'''
assert original_object.count(needle) == 1
zero_fields = needle + '''
		for index, binding := range lowered.Body {
			if declared, known := binding.(ir.Declare); known && declared.Value.Type() == ir.Number {
				declared.Value = ir.NumberConstant{Value: 0}
				lowered.Body[index] = declared
			}
		}'''
repeat_source = needle + '\n\t\tlowered.Body = append([]ir.Statement{ir.Evaluate{Value: iterable}}, lowered.Body...)'
mutants = [
    ('omit-getter-dispatch', accessor_source,
     original_accessor.replace('if names[expression.Name] {', 'if names[expression.Name] && expression.Name != "weight" {', 1),
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure_sites',
     'stdout differs'),
    ('reject-object-binding', object_source,
     original_object.replace(' && name.Kind != ast.KindObjectBindingPattern', '', 1),
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure_sites',
     'a for...of destructuring an object'),
    ('zero-field-binding', object_source, original_object.replace(needle, zero_fields),
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure_sites',
     'stdout differs'),
    ('repeat-source', object_source, original_object.replace(needle, repeat_source),
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure_sites',
     'stdout differs'),
    ('omit-close', iteration_source,
     original_iteration.replace('if state.plan.close == nil {', 'if true {', 1),
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure_iterable',
     'stdout differs'),
    ('wrong-close-receiver', iteration_source,
     original_iteration.replace('invokeMember(function, receiver, direct, state.iterator, nil, ir.Object)',
                                'invokeMember(function, receiver, direct, ir.ObjectLiteral{Fields: []ir.Field{{Name: "index", Value: ir.NumberConstant{Value: 0}}}}, nil, ir.Object)', 1),
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_object_destructure_iterable',
     'stdout differs'),
]
environment = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
for name, path, changed, package, test, catcher in mutants:
    before = path.read_text()
    assert changed != before, name
    try:
        path.write_text(changed)
        log = logs / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m'],
                                    cwd=repository, env=environment, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and catcher in observed, (name, observed)
        assert '[-Werror' not in observed and '[build failed]' not in observed, (name, observed)
        print(f'{name}: caught ({catcher}), exit {result.returncode}', flush=True)
    finally:
        path.write_text(before)
