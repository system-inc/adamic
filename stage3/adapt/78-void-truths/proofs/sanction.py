#!/usr/bin/env python3
"""Extend the independently reconstructed main sanction by exactly System.exit."""
import difflib
import json
import os
from pathlib import Path
import subprocess
import sys

repo, upstream, output = map(Path, sys.argv[1:])
environment = dict(os.environ, NODE_PATH=str(repo / 'stage3/api/node_modules'))
proof = json.loads(subprocess.check_output(['node', str(repo / 'stage3/lane/generate-api.cjs'), str(upstream)], env=environment, timeout=60))
original = proof.pop('original')
previous = proof.pop('expected')
old = 'exit(exitCode?: number): void;'
new = 'exit(exitCode?: number): number | undefined;'
assert previous.count(old) == 1
expected = previous.replace(old, new)
original_file = output.parent / 'original-api.d.ts.txt'
expected_file = output.parent / 'expected-api.d.ts.txt'
original_file.write_text(original)
expected_file.write_text(expected)
proof['normalized_changes'] = json.loads(subprocess.check_output(['node', str(repo / 'stage3/lane/normalize-api.cjs'), str(original_file), str(expected_file)], env=environment, timeout=30))
proof['counts']['78'] = 1
proof['owners'].append(dict(adaptation='78', path=['System', 'exit'], line=original[:original.index(old)].count('\n') + 1))
proof['baseline_diffs'] = ['api/typescript.d.ts']
proof['diff_lines'] = list(difflib.unified_diff(original.splitlines(True), expected.splitlines(True), fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts'))
proof['ruled_additions'] = dict(task='#se1zcgr', removed=old, added=new, declaration='System.exit', previous_declarations=len(proof['normalized_changes']) - 1)
output.write_text(json.dumps(proof, indent=2) + '\n')
(output.parent / 'api-addition.diff').write_text(''.join(difflib.unified_diff(previous.splitlines(True), expected.splitlines(True), fromfile='main-sanction/api/typescript.d.ts', tofile='adapt78-sanction/api/typescript.d.ts')))
print(json.dumps(dict(status='pass', previous_declarations=len(proof['normalized_changes']) - 1, added_declarations=1, output=str(output))))
