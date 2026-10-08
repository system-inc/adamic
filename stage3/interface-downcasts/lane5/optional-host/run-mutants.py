#!/usr/bin/env python3
"""Run each guard omission independently; always restore production sources."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[4]
os.chdir(root)
logs = Path(os.environ.get('ADAMIC_OPTIONAL_HOST_LOGS', '/tmp/views-optional-host-mutants'))
logs.mkdir(parents=True, exist_ok=True)
mutants = [
    ('native-presence', 'internal/native/runtime/view_nullish.c',
     'if(slot==NULL&&absent)return NULL;', 'if(absent)return NULL;', 'good'),
    ('javascript-presence', 'internal/javascript/readiness.go',
     'if (allowAbsent && !Object.hasOwn(object, name)) return undefined;',
     'if (allowAbsent) return undefined;', 'good'),
    ('native-callable', 'internal/native/runtime/view_callables_contract.h',
     'if (optional && value == NULL) { return NULL; }',
     'if (optional && value == NULL) { return NULL; }\n    if (value != NULL && value->kind == adamic_kind_closure) return value;', 'wrong-signature-read'),
    ('javascript-callable', 'internal/javascript/view_callables_contract.go',
     'if (optional && value === undefined) return undefined;',
     'if (optional && value === undefined) return undefined;\n    if (adamicTypeOf(value) === "function") return value;', 'wrong-signature-read'),
    ('plain-error-invention', 'internal/lower/cast_proof.go',
     'if l.isLibraryType(source, "Error") {', 'if false && l.isLibraryType(source, "Error") {', None),
]
for name, filename, original, changed, variant in mutants:
    path = root / filename
    content = path.read_text()
    assert content.count(original) == 1, (name, 'mutation anchor changed')
    log = logs / (name + '.log')
    if variant:
        command = ['go', 'test', './internal/oracle', '-run',
                   '^TestCheckedViewOptionalHostMethods$/.*/' + variant + '$', '-count=1', '-v']
        pin = 'callable pin:' if variant == 'wrong-signature-read' else 'got oracle stdout'
    else:
        command = ['go', 'test', './internal/lower', '-run',
                   '^TestOptionalHostMethodBoundaries$/invented_Error', '-count=1', '-v']
        pin = 'got <nil>'
    try:
        path.write_text(content.replace(original, changed))
        with log.open('w') as output:
            status = subprocess.run(command, stdout=output, stderr=subprocess.STDOUT).returncode
        result = log.read_text()
        assert status != 0 and '--- FAIL: Test' in result and '[build failed]' not in result, (name, status, result)
        if variant == 'good':
            assert 'stdout differs' in result and 'exit=0' in result, (name, result)
        else:
            assert pin in result, (name, result)
            if variant: assert 'exit=0' in result and 'AddressSanitizer' not in result, (name, result)
        print(f'{name}: killed by semantic assertion; exit={status}; log={log}', flush=True)
    finally:
        path.write_text(content)
