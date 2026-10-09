#!/usr/bin/env python3
"""Prove debugger artifact, Node, preservation and flow checks can fail."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
logs = Path('/tmp/adamic-debugger-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('trap-artifact', 'internal/native/emit_statements.go',
     'case ir.Debugger:', 'case ir.Debugger:\n\t\te.line("__builtin_trap();")',
     './internal/oracle', 'TestDebuggerNativeEmitsNothing', 'debugger emitted a trap or breakpoint call'),
    ('trap-oracle', 'internal/native/emit_statements.go',
     'case ir.Debugger:', 'case ir.Debugger:\n\t\te.line("__builtin_trap();")',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/debugger_fail', 'exit codes differ'),
    ('javascript-dropped', 'internal/javascript/javascript.go',
     'e.line("debugger;")', '',
     './internal/oracle', 'TestDebuggerNativeEmitsNothing', 'JavaScript kept 0 debugger statements'),
    ('lowering-dropped', 'internal/lower/statements.go',
     'return []ir.Statement{ir.Debugger{}}, nil', 'return nil, nil',
     './internal/oracle', 'TestDebuggerNativeEmitsNothing', 'JavaScript kept 0 debugger statements'),
    ('debugger-refused', 'internal/lower/refusals.go',
     'var refusals = map[ast.Kind]refusal{',
     'var refusals = map[ast.Kind]refusal{\n\tast.KindDebuggerStatement: {"debugger", "remove it"},',
     './internal/oracle', 'TestDebuggerNativeEmitsNothing', 'refuses debugger'),
    ('flow-instruction', 'internal/flow/build.go',
     '// With no debugger attached, it reads and writes nothing and keeps the path.\n\t\treturn',
     '// With no debugger attached, it reads and writes nothing and keeps the path.\n\t\tb.emit(at, 0, nil, nil, nil)\n\t\treturn',
     './internal/flow', 'TestDebuggerHasNoFlowInstruction', 'debugger changed the flow: 2 instructions'),
]
for name, file, old, new, package, test, catcher in mutants:
    path = root / file
    original = path.read_text()
    assert original.count(old) == 1, name
    try:
        path.write_text(original.replace(old, new))
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', '-count=1', '-timeout', '10m', package, '-run', test], cwd=root, stdout=log, stderr=subprocess.STDOUT, env=os.environ | {'ADAMIC_GATE_UNCACHED': '1'})
        output = (logs / (name + '.log')).read_text()
        assert result.returncode != 0 and catcher in output, f'{name} not caught: {output}'
        assert 'build failed' not in output and 'clang failed' not in output, output
        print(f'{name}: caught by {catcher}, exit {result.returncode}', flush=True)
    finally:
        path.write_text(original)
