#!/usr/bin/env python3
"""Prove the optional read, required read and exception analysis guards.

Run after sourcing cloud/setup.sh's environment file. Each mutant runs alone;
the exact original file is restored even when a command fails.
"""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
logs = Path('/tmp/optional-after-call-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('throwing-optional-receiver', 'internal/lower/narrowed.go',
     'l.acceptsUndefined(node) || optionalReceiver(node)',
     'comparedWithUndefined(node)', '79'),
    ('required-read-panics', 'internal/ir/defined.go',
     'return strings.HasPrefix(d.Message, "TypeError: ")',
     'return strings.HasPrefix(d.Message, "mutant: ")', 'required'),
    ('missing-throw-propagation', 'internal/lower/exceptions.go',
     'case ir.Defined:\n\t\t\tfound = node.Throws()',
     'case ir.Defined:\n\t\t\tfound = false && node.Throws()', 'propagation'),
    ('missing-exception-edge', 'internal/flow/build.go',
     'if value.Interface().(ir.Defined).Throws() {',
     'if false && value.Interface().(ir.Defined).Throws() {', 'propagation'),
]
for name, filename, before, after, fixture in mutants:
    path = root / filename
    original = path.read_text()
    assert original.count(before) == 1, name
    log = logs / (name + '.log')
    try:
        path.write_text(original.replace(before, after, 1))
        selector = 'TestNativeAgreesWithNode/internal/oracle/testdata/optional_after_call_' + fixture + r'\.a$'
        package = './internal/oracle'
        if name == 'missing-exception-edge':
            package, selector = './internal/flow', 'TestDefinedExceptionEdges'
        with log.open('wb') as output:
            result = subprocess.run(['go', 'test', package, '-run', selector,
                                     '-count=1', '-timeout', '30m'], cwd=root,
                                    env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'},
                                    stdout=output, stderr=subprocess.STDOUT)
    finally:
        path.write_text(original)
    output = log.read_text()
    comparison = result.returncode != 0 and 'node:' in output and any(marker in output for marker in ['exit codes differ', 'stdout differs', 'stderr differs'])
    sanitizer = result.returncode != 0 and 'AddressSanitizer: heap-use-after-free' in output
    unrelated = any(marker in output for marker in ['build failed', '[build failed]',
                      'compiling runtime', 'undefined reference', 'error: unused'])
    edge = result.returncode != 0 and 'catch successor: got false, want true' in output
    caught = (comparison or sanitizer or edge) and not unrelated
    print(f'{name}: {"CAUGHT" if caught else "INVALID"}; exit={result.returncode}; log={log}', flush=True)
    if not caught:
        raise SystemExit('Mutant was not caught by execution: ' + name)
