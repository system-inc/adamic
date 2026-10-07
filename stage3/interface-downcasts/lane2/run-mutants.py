#!/usr/bin/env python3
"""Run independent semantic mutants of the currently standalone lane 2 hooks."""
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path('/tmp/adamic-view-lane2-mutants')
logs.mkdir(exist_ok=True)
cases = [
    ('array-kind', 'internal/javascript/view_arrays.go',
     'if (!Array.isArray(value))', 'if (false)',
     './internal/javascript', 'TestViewArraysNode/kind', 'TestViewArraysNode/kind'),
    ('element', 'internal/javascript/view_arrays.go',
     'return check(value, expression + "[" + index + "]");', 'return value;',
     './internal/javascript', 'TestViewArraysNode/second', 'TestViewArraysNode/second'),
    ('callable-kind', 'internal/javascript/view_callables.go',
     'if (adamicTypeOf(value) !== "function")', 'if (false)',
     './internal/javascript', 'TestViewCallablesNode/kind', 'TestViewCallablesNode/kind'),
    ('signature', 'internal/lower/view_callables.go',
     'if signatureProven {', 'if true {',
     './internal/lower', 'TestViewCallableSignature', 'TestViewCallableSignature'),
]
for name, relative, before, after, package, pattern, failure in cases:
    path = root / relative
    original = path.read_bytes()
    text = original.decode()
    assert text.count(before) == 1, name
    try:
        path.write_text(text.replace(before, after))
        log = logs / (name + '.log')
        with log.open('wb') as output:
            result = subprocess.run(['go', 'test', package, '-run', pattern,
                                     '-count=1', '-timeout', '10m'],
                                    cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and ('--- FAIL: ' + failure) in observed, observed
        assert '[build failed]' not in observed, observed
        print(name + ': caught by ' + failure + '; ' + str(log), flush=True)
    finally:
        path.write_bytes(original)
