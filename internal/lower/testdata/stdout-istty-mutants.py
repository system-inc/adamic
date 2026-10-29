#!/usr/bin/env python3
"""Each isTTY boundary has a focused red test; overlays never change the tree."""
import json
import pathlib
import subprocess
import tempfile

root = pathlib.Path(__file__).resolve().parents[3]
cases = [
    ('false-off-tty', 'internal/native/library_node_fs_file.go',
     '(adamic_maybe_boolean){isatty(1) != 0, true}',
     '(adamic_maybe_boolean){true, isatty(1) != 0}',
     './internal/oracle', 'TestStdoutIsTTYMatchesNodeOnPipe', 'backend 1, pipe: stdout differs'),
    ('admit-stderr', 'internal/lower/library_node_host.go',
     'l.nodeLibraryMember(node.AsPropertyAccessExpression().Expression) == "node:process.Process.stdout"\n}',
     '(l.nodeLibraryMember(node.AsPropertyAccessExpression().Expression) == "node:process.Process.stdout" || l.nodeLibraryMember(node.AsPropertyAccessExpression().Expression) == "node:process.Process.stderr")\n}',
     './internal/lower', 'TestNodeStdoutIsTTYRefusesOtherStreams', 'want "node:process.Process.stderr.isTTY'),
    ('admit-write', 'internal/lower/library_node_host.go',
     'if ast.IsAssignmentTarget(outer) {\n\t\treturn l.notYet(node, name+" written',
     'if false && ast.IsAssignmentTarget(outer) {\n\t\treturn l.notYet(node, name+" written',
     './internal/lower', 'TestNodeStdoutIsTTYRefusesMutation', 'written (fix: read process.stdout.isTTY; terminal status is read-only)'),
]
for name, file, before, after, package, test, reason in cases:
    source = root / file
    text = source.read_text()
    # The receiver spelling also occurs in write's recognizer. Restrict this
    # mutant to the isTTY function, leaving stdout.write's admission unchanged.
    if name == 'admit-stderr':
        start = text.index('func (l *lowering) nodeStdoutIsTTYMember(')
        end = text.index('\nfunc ', start + 1)
    else:
        start, end = 0, len(text)
    part = text[start:end]
    if part.count(before) != 1:
        raise SystemExit(f'{name}: mutation must match exactly once')
    mutated = text[:start] + part.replace(before, after, 1) + text[end:]
    with tempfile.TemporaryDirectory(prefix='stdout-istty-mutant-') as tmp:
        replacement = pathlib.Path(tmp) / source.name
        replacement.write_text(mutated)
        overlay = pathlib.Path(tmp) / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
        command = ['go', 'test', '-json', '-overlay', str(overlay), package,
                   '-count=1', '-p', '1', '-parallel', '1', '-timeout', '2m', '-run', '^' + test + '$']
        result = subprocess.run(command, cwd=root, stdout=subprocess.PIPE,
                                stderr=subprocess.STDOUT, timeout=180, text=True)
        print(result.stdout, end='', flush=True)
        events = []
        for line in result.stdout.splitlines():
            try:
                events.append(json.loads(line))
            except ValueError:
                pass
        output = ''.join(row.get('Output', '') for row in events)
        red = any(row.get('Test') == test and row.get('Action') == 'fail' for row in events)
        if result.returncode != 1 or reason not in output or not red:
            raise SystemExit(f'{name}: {test} did not catch its boundary mutant')
        print(f'caught: {name}: {test}', flush=True)
