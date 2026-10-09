#!/usr/bin/env python3
"""Each mutant starts from the final compiler and is restored after its test."""
from pathlib import Path
import subprocess
import os

root = Path(__file__).resolve().parents[4]
logs = Path('/tmp/adamic-method-presence-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('shadowed-undefined', 'internal/lower/method_presence.go',
     'len(symbol.Declarations) == 0 &&',
     '',
     './internal/lower', 'TestMethodPresenceRefusesEscapes', 'got <nil>, want unbound-method'),
    ('stored', 'internal/lower/method_presence.go',
     'switch parent.Kind {',
     'switch parent.Kind {\n\tcase ast.KindVariableDeclaration:\n\t\treturn true',
     './internal/lower', 'TestMethodPresenceRefusesEscapes', 'got <nil>, want unbound-method'),
    ('prototype-absent', 'internal/native/method_presence.go',
     'e.line("\\t\\t%s = true;", present)',
     'e.line("\\t\\t%s = false;", present)',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_presence', 'stdout'),
    ('absent-present', 'internal/native/method_presence.go',
     'e.line("bool %s = false;", present)',
     'e.line("bool %s = true;", present)',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_presence', 'stdout'),
    ('wide-name-omitted', 'internal/native/emit_objects.go',
     'thunks = append(thunks, "NULL")',
     'methodNames = methodNames[:len(methodNames)-1]',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_presence', 'stdout'),
]
for name, file, old, new, package, test, catcher in mutants:
    path = root / file
    original = path.read_text()
    assert original.count(old) == 1, name
    try:
        path.write_text(original.replace(old, new))
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '10m'], cwd=root, stdout=log, stderr=subprocess.STDOUT, env=os.environ | {'ADAMIC_GATE_UNCACHED': '1'})
        output = (logs / (name + '.log')).read_text()
        assert result.returncode != 0 and catcher in output, f'{name} not caught by {catcher}: {output}'
        assert 'clang failed' not in output, f'{name} killed by compiler warnings'
        print(f'{name}: caught by {catcher}, exit {result.returncode}', flush=True)
    finally:
        path.write_text(original)
