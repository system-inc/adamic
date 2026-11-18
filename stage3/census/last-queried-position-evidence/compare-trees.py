import os
from pathlib import Path
import subprocess
import sys

before, after = map(Path, sys.argv[1:])
parser = b'src/compiler/parser.ts'

def snapshot(tree):
    raw = subprocess.check_output(['git', '-C', tree, 'ls-tree', '-rz', 'refs/stage3/99-wasm-crypto-import-after'])
    return dict(entry.split(b'\t', 1)[::-1] for entry in raw.split(b'\0') if entry)

old, new = snapshot(before), snapshot(after)
assert set(old) == set(new), 'applied paths differ'
changed, = (path for path in old if old[path] != new[path])
assert changed == parser, os.fsdecode(changed)
previous = (before / os.fsdecode(parser)).read_bytes()
candidate = (after / os.fsdecode(parser)).read_bytes()
original = b'let lastQueriedPosition = InvalidPosition.Value;'
assert previous.count(original) == 1
assert candidate == previous.replace(original, b'let lastQueriedPosition: number = InvalidPosition.Value;')
print('Complete adapted trees differ only by the parser annotation.')
