#!/usr/bin/env python3
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/lane5-ranked-mutants-' + '-'.join(__import__('sys').argv[1:]))
logs.mkdir(exist_ok=True)
selection = '^TestCheckedViewCallableRankedFamilies$'
groups = __import__('sys').argv[1:]
assert groups
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
if all(group in ['emit-notification', 'substitution'] for group in groups):
    mutants = [mutant for mutant in mutants if mutant[4] in ['wrong-arity', 'wrong-members']]
for name, filename, before, after, variant in mutants:
    path = root / filename
    original = path.read_text()
    assert original.count(before) == 1
    try:
        path.write_text(original.replace(before, after))
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', './internal/oracle', '-run', selection + '/^(' + '|'.join(groups) + ')$/^' + variant + '$', '-count=1', '-timeout', '5m'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        evidence = (logs / (name + '.log')).read_text()
        assert result.returncode != 0 and '--- FAIL:' in evidence, evidence
        assert '[build failed]' not in evidence and 'clang failed:' not in evidence, evidence
        assert 'backend' in evidence, evidence
        if name == 'payload-reads':
            assert 'AddressSanitizer: SEGV' in evidence, evidence
        for group in groups:
            assert group + '/' in evidence, evidence
        print(name + ': caught for selected original member fixtures', flush=True)
    finally:
        path.write_text(original)

