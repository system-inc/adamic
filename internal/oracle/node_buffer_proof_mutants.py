#!/usr/bin/env python3
"""Prove the cycle and runtime-layout guards, independently of Node semantics."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[2]
logs = Path('/tmp/host-buffer-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('hash_alias', 'internal/fresh/library_node_buffer.go', 'return receiver', '_ = receiver; return a.fresh(anyField, value{})', './internal/fresh', 'TestNodeBufferHashUpdateKeepsAlias', 'Hash.update lost its receiver alias'),
    ('hash_fields', 'internal/native/fields.go', '\n\t\t{"bytes", "finalized"},', '', './internal/native', 'TestRuntimeFieldLayoutsAreIncluded', 'runtime field "bytes" at 0 is absent'),
    ('host_cycle_dispatch', 'internal/fresh/library_language.go', '\n\tcase ir.NodeBufferCall:\n\t\treturn a.nodeBufferCall(expression), true', '', './internal/fresh', 'TestEveryWriteIsRecordedAndKnown', 'ir.NodeBufferCall is a node the cycle finder'),
]
summary = []
for name, filename, before, after, package, test, expected in mutants:
    path = root / filename
    original = path.read_text()
    assert before in original, name
    try:
        path.write_text(original.replace(before, after, 1))
        logfile = logs / (name + '.log')
        with logfile.open('wb') as output:
            result = subprocess.run(['go', 'test', package, '-run', '^' + test + '$', '-count=1', '-timeout', '30m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        output = logfile.read_text()
        assert result.returncode != 0 and expected in output and 'build failed' not in output, name
        line = name + ': CAUGHT by ' + test + '; log=' + str(logfile)
        print(line, flush=True)
        summary.append(line)
    finally:
        path.write_text(original)
(logs / 'proof-summary.log').write_text('\n'.join(summary) + '\n')
