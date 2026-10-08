#!/usr/bin/env python3
"""Overlay only each analysis hook; require an oracle or sanitizer failure after a clean build."""
import json
import pathlib
import subprocess
import tempfile

root = pathlib.Path(__file__).resolve().parents[3]
mutants = [
    ('global-lending', 'internal/native/reuse.go', 'if expression.Replacement != nil {\n\t\t\t\t\t\tfound = true', 'if expression.Replacement != nil && false {\n\t\t\t\t\t\tfound = true', 'effects'),
    ('global-move', 'internal/native/reuse.go', 'if expression.Replacement != nil {\n\t\t\t\treached = true', 'if expression.Replacement != nil && false {\n\t\t\t\treached = true', 'move_effect'),
    ('exception-edge', 'internal/flow/build.go', 'if value.Interface().(ir.RegExpCall).Replacement != nil && program.ClosuresMayThrow {', 'if value.Interface().(ir.RegExpCall).Replacement != nil && false {', 'effects'),
]
logs = pathlib.Path('/tmp/regex-replace-effects-mutants')
logs.mkdir(exist_ok=True)
for name, file, old, new, fixture in mutants:
    original = root / file
    source = original.read_text()
    if source.count(old) != 1:
        raise SystemExit(f'{name}: mutation site moved')
    with tempfile.TemporaryDirectory(prefix='regexp-replacement-mutant-') as scratch:
        directory = pathlib.Path(scratch)
        replacement = directory / original.name
        replacement.write_text(source.replace(old, new, 1))
        overlay = directory / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(original): str(replacement)}}))
        command = ['go', 'test', '-overlay', str(overlay), './internal/oracle', '-run', f'^TestNativeAgreesWithNode/internal/oracle/testdata/regexp_replace/{fixture}.a$', '-count=1', '-v']
        with (logs / (name + '.log')).open('wb') as output:
            result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT)
        output = (logs / (name + '.log')).read_text()
        catches = [marker for marker in ('heap-use-after-free', 'stdout differs', 'exit codes differ', 'runtime error:') if marker in output]
        if result.returncode == 0 or 'build failed' in output or not catches:
            raise SystemExit(f'{name}: did not reach the intended oracle/sanitizer failure; see {logs / (name + ".log")}')
        print(name + ': ' + ', '.join(catches))
