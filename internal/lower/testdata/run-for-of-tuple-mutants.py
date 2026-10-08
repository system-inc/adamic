#!/usr/bin/env python3
"""Prove tuple storage and iteration rules with isolated, restored mutants."""
from pathlib import Path
import os
import subprocess

repository = Path(__file__).resolve().parents[3]
source = repository / 'internal/lower/for_of_tuple.go'
original = source.read_text()
logs = Path('/tmp/adamic-for-of-tuple-mutants')
logs.mkdir(exist_ok=True)
oracle = 'TestNativeAgreesWithNode/internal/oracle/testdata/for_of_tuple'
guards = 'TestForOfTupleStorageChecks'
mutants = [
    ('empty-storage', original.replace(' || len(elements) == 0', ''), './internal/lower', guards, 'want tuple storage stop'),
    ('pattern-storage', original.replace('if !ast.IsIdentifier(name) {', 'if !ast.IsIdentifier(name) && false {'), './internal/lower', guards, 'want tuple storage stop'),
    ('optional-storage', original.replace('checker.TupleType_combinedFlags(proven.TargetTupleType())&checker.ElementFlagsNonRequired != 0 || ', ''), './internal/lower', guards, 'want tuple storage stop'),
    ('mixed-storage', original.replace('!known || stored != of || (of != ir.String && of != ir.Number && of != ir.Boolean)', '(!known || stored != of || (of != ir.String && of != ir.Number && of != ir.Boolean)) && false'), './internal/lower', guards, 'want tuple storage stop'),
    ('short-length', original.replace('float64(len(elements))', 'float64(len(elements) - 1)'), './internal/oracle', oracle, 'stdout differs'),
    ('first-field', original.replace('strconv.Itoa(slot)', 'strconv.Itoa(0)').replace('strconv.Itoa(len(elements) - 1)', 'strconv.Itoa(0)'), './internal/oracle', oracle, 'stdout differs'),
    ('repeat-source', original.replace('Body:      append([]ir.Statement{', 'Body:      append([]ir.Statement{ir.Evaluate{Value: source}, '), './internal/oracle', oracle, 'stdout differs'),
    ('shared-binding', original.replace('Body:      append([]ir.Statement{ir.Declare{Local: local, Value: value}}', 'Body:      append([]ir.Statement{ir.Assign{Local: local, Value: value}}').replace('Value: ir.NumberConstant{Value: 0}}, loop', 'Value: ir.NumberConstant{Value: 0}}, ir.Declare{Local: local, Value: value}, loop'), './internal/oracle', oracle, 'stdout differs'),
]
for name, changed, package, test, catcher in mutants:
    assert changed != original, name
    try:
        source.write_text(changed)
        log = logs / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m'], cwd=repository, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and catcher in observed, (name, observed)
        assert '[build failed]' not in observed and '[-Werror' not in observed, (name, observed)
        print(f'{name}: caught ({catcher}), exit {result.returncode}', flush=True)
    finally:
        source.write_text(original)
