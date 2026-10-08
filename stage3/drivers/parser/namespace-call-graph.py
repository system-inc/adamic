#!/usr/bin/env python3
"""Bound namespace call-graph probes; never count a timeout as a refusal."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time

p = argparse.ArgumentParser()
p.add_argument('compiler', type=Path)
p.add_argument('scratch', type=Path)
p.add_argument('output', type=Path)
a = p.parse_args()
a.compiler = a.compiler.resolve()
a.scratch = a.scratch.resolve()
a.output = a.output.resolve()
a.output.mkdir(exist_ok=False)
rows = []
for depth, edges in [(12, 2), (16, 2), (20, 2), (24, 2), (24, 1)]:
    label = f'depth-{depth}-edges-{edges}'
    entry = a.output / (label + '.a')
    body = 'function f0(): void {}\n'
    for i in range(1, depth + 1):
        body += f'function f{i}(): void {{ if (false) {{ ' + f'f{i - 1}(); ' * edges + '} }\n'
    body += f'f{depth}();\nnamespace Debug {{ export let ready = true; }}\nconsole.log(`${{Debug.ready}}`);\n'
    entry.write_text(body)
    row = {'depth': depth, 'edges': edges}
    node = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(entry)], cwd=a.scratch, capture_output=True)
    row['node'] = {'exit': node.returncode, 'stdout': node.stdout.decode(), 'stderr': node.stderr.decode()}
    start = time.monotonic()
    with (a.output / (label + '.c')).open('wb') as stdout, (a.output / (label + '.stderr')).open('wb') as stderr:
        try:
            native = subprocess.run([str(a.compiler), 'c', str(entry)], cwd=a.scratch, stdout=stdout, stderr=stderr, timeout=30)
            row['emit_exit'] = native.returncode
            row['timeout'] = False
        except subprocess.TimeoutExpired:
            row['emit_exit'] = None
            row['timeout'] = True
    row['wall_seconds'] = time.monotonic() - start
    rows.append(row)
    (a.output / 'report.json').write_text(json.dumps(rows, indent=2) + '\n')
print(json.dumps(rows, indent=2))
