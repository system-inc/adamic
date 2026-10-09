#!/usr/bin/env python3
"""Patch only the pinned scratch checkout, preserving its source bytes elsewhere."""
from pathlib import Path
import hashlib
import subprocess
import sys

root = Path(sys.argv[1]).resolve()
pin = '050880ce59e30b356b686bd3144efe24f875ebc8'
assert subprocess.check_output(['git', '-C', str(root), 'rev-parse', 'HEAD'], text=True).strip() == pin
path = root / 'src/compiler/utilities.ts'
raw = path.read_bytes()
assert raw == subprocess.check_output(['git', '-C', str(root), 'show', 'HEAD:src/compiler/utilities.ts'])
for statement, hook in [
    ('(node as Mutable<T>).flags = newFlags;', 'flags", node, newFlags, node'),
    ('(visited as Mutable<T>).parent = undefined!;', 'parent", visited, undefined, node'),
]:
    needle = statement.encode()
    assert raw.count(needle) == 1
    raw = raw.replace(needle, needle + b'\r\n' + ('        (globalThis as any).__twoWrites.write("' + hook + ');').encode())
path.write_bytes(raw)
print('instrumented', path, hashlib.sha256(raw).hexdigest())
