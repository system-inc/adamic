"""Run restored guard mutants; each must fail its targeted checked-type probe."""
import json
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
helper = root / 'internal/lower/representation_clock_children_optional.go'
expression = root / 'internal/lower/expression.go'
mutants = [
    ('clock-children-missing-constraint-guard', helper,
     'constraint == nil || seen[constraint]', 'seen[constraint]',
     'TestClockChildrenOptionalRepresentation/unconstrained'),
    ('clock-children-noninterface-is-array', helper,
     'constraint.ObjectFlags()&checker.ObjectFlagsInterface == 0 {\n\t\treturn false',
     'constraint.ObjectFlags()&checker.ObjectFlagsInterface == 0 {\n\t\treturn true',
     'TestClockChildrenOptionalRepresentation/callable_constraint'),
    ('clock-children-forget-reference-target', helper,
     'constraint = constraint.Target()', '// omitted reference target',
     'TestClockChildrenOptionalRepresentation/array_ancestry'),
    ('clock-children-structural-is-array', helper,
     'seen[constraint] = true', 'seen[constraint] = true\n return true',
     'TestClockChildrenOptionalRepresentation/structural_length'),
    ('clock-children-tuple-is-array', helper,
     'if constraint == nil || seen[constraint] || checker.IsTupleType(constraint) {',
     'if constraint != nil && checker.IsTupleType(constraint) { return true }; if constraint == nil || seen[constraint] {',
     'TestClockChildrenOptionalRepresentation/tuple'),
    ('clock-children-forget-array-ancestry', helper,
     'for _, base := range l.checker.GetBaseTypes(constraint) {',
     'return false\n for _, base := range l.checker.GetBaseTypes(constraint) {',
     'TestClockChildrenOptionalRepresentation/array_ancestry'),
    ('clock-children-ignore-substitution', expression,
     'return substituted, true', 'return ir.Array, substituted == substituted',
     'TestClockChildrenOptionalRepresentation/readonly_array'),
    ('clock-children-forget-array-proof', helper,
     'return true\n\t}', 'return false\n\t}',
     'TestClockChildrenOptionalRepresentation/readonly_array'),
]
results = []
for name, file, before, after, test in mutants:
    original = file.read_text()
    if original.count(before) != 1:
        raise RuntimeError(f'{name}: expected one mutation site, got {original.count(before)}')
    try:
        file.write_text(original.replace(before, after, 1))
        log = Path('/tmp') / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', './internal/lower', '-count=1', '-run', '^' + test + '$', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        content = log.read_text()
        killed = result.returncode != 0 and '--- FAIL: TestClockChildrenOptionalRepresentation' in content
        results.append(dict(name=name, test=test, exit=result.returncode, killed=killed, log=str(log)))
        if not killed:
            raise RuntimeError(f'{name} survived or failed to compile; see {log}')
    finally:
        file.write_text(original)
print(json.dumps(results, indent=2))
