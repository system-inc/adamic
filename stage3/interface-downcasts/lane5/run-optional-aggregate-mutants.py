#!/usr/bin/env python3
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/lane5-optional-aggregate-mutants')
logs.mkdir(exist_ok=True)
selection = '^TestCheckedViewCallableOptionalAggregates$'
mutants = [
    ('native-arity', 'internal/native/runtime/view_callables_contract.h',
     '        if (recorded->arity != expected->arity) {',
     '        if (recorded->arity != expected->arity) return value;\n        if (recorded->arity != expected->arity) {', 'wrong-arity'),
    ('javascript-arity', 'internal/javascript/view_callables_contract.go',
     '            if (recorded.parameters.length !== expected.parameters.length) found',
     '            if (recorded.parameters.length !== expected.parameters.length) return value;\n            if (recorded.parameters.length !== expected.parameters.length) found', 'wrong-arity'),
    ('native-result', 'internal/native/runtime/view_callables_contract.h',
     '} else if (expected->result != 255 && !adamic_callable_representation_compatible',
     '} else if (false && expected->result != 255 && !adamic_callable_representation_compatible', 'wrong-result'),
    ('javascript-result', 'internal/javascript/view_callables_contract.go',
     'else if (expected.result !== 255 && !adamicCallableRepresentationCompatible',
     'else if (false && expected.result !== 255 && !adamicCallableRepresentationCompatible', 'wrong-result'),
    ('native-parameters', 'internal/native/runtime/view_callables_contract.h',
     '            found = "function with incompatible parameter representations";',
     '            return value;', 'wrong-members'),
    ('javascript-parameters', 'internal/javascript/view_callables_contract.go',
     '            else found = "function with incompatible parameter representations";',
     '            else return value;', 'wrong-members'),
    ('payload-reads', 'internal/lower/view_callables_aggregate.go',
     'program.CheckedFields[field.Name] = true',
     'continue // mutant: omit aggregate descendant read registration', 'wrong-parameter-payload'),
]
for name, filename, before, after, variant in mutants:
    path = root / filename
    original = path.read_text()
    assert original.count(before) == 1
    try:
        path.write_text(original.replace(before, after))
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', './internal/oracle', '-run', selection + '/.*/^' + variant + '$', '-count=1', '-timeout', '5m'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        evidence = (logs / (name + '.log')).read_text()
        assert result.returncode != 0 and '--- FAIL:' in evidence, evidence
        assert '[build failed]' not in evidence and 'clang failed:' not in evidence, evidence
        assert 'backend' in evidence, evidence
        if name == 'payload-reads':
            assert 'AddressSanitizer: SEGV' in evidence, evidence
        for group in ['parameter-declaration/', 'variable-statement/']:
            assert group in evidence, evidence
        print(name + ': caught for both original member fixtures', flush=True)
    finally:
        path.write_text(original)

path = root / 'internal/lower/view_callables.go'
original = path.read_text()
before = 'name, named := viewCallableImplementationName(part)'
assert original.count(before) == 1
try:
    path.write_text(original.replace(before, 'name, named := part.Name(), true'))
    with (logs / 'implementation-names.log').open('w') as log:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^TestCheckedViewCallableDestructuredSiblingRefusal$', '-count=1', '-timeout', '5m'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
    evidence = (logs / 'implementation-names.log').read_text()
    assert result.returncode != 0 and 'panic: Unhandled case in Node.Text: *ast.BindingPattern' in evidence, evidence
    assert '[build failed]' not in evidence and 'clang failed:' not in evidence, evidence
    print('implementation-names: restored AST panic caught by Node-backed refusal regression', flush=True)
finally:
    path.write_text(original)
