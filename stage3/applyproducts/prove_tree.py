#!/usr/bin/env python3
"""Compare a fresh main apply with product 99, including every non-.git entry."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import time

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'stage3'))


def tree_hash(root):
    paths = []
    for directory, directories, files in os.walk(root, followlinks=False):
        if Path(directory) == root:
            directories[:] = [name for name in directories if name != '.git']
        paths += [Path(directory) / name for name in directories + files]
    digest = hashlib.sha256()
    for path in sorted(paths, key=lambda path: path.relative_to(root).as_posix()):
        mode = path.lstat().st_mode
        kind = 'link' if stat.S_ISLNK(mode) else 'directory' if stat.S_ISDIR(mode) else 'file'
        data = os.readlink(path).encode() if kind == 'link' else b'' if kind == 'directory' else path.read_bytes()
        metadata = json.dumps([path.relative_to(root).as_posix(), kind, stat.S_IMODE(mode)]).encode()
        for part in (metadata, data):
            digest.update(len(part).to_bytes(8, 'big'))
            digest.update(part)
    return digest.hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('reference', type=Path)
    parser.add_argument('product', type=Path)
    parser.add_argument('--main', default='origin/main')
    args = parser.parse_args()
    original = subprocess.check_output(['git', 'show', args.main + ':stage3/apply.py'], cwd=ROOT)
    namespace = dict(__file__=str(ROOT / 'stage3/apply.py'), __name__='plain_main_apply')
    exec(compile(original, namespace['__file__'], 'exec'), namespace)
    started = time.monotonic()
    namespace['build_product'](args.reference)
    print(f'plain main apply seconds={time.monotonic() - started:.6f}', flush=True)
    baseline, product = tree_hash(args.reference), tree_hash(args.product / 'tree')
    print(json.dumps(dict(main=baseline, product99=product)), flush=True)
    assert baseline == product, 'whole-tree equality failed'
    target = args.reference / 'src/compiler/core.ts'
    before = target.read_bytes()
    try:
        target.write_bytes(before + b'\n// equality mutant\n')
        assert tree_hash(args.reference) != product, 'whole-tree comparison accepted changed bytes'
        print('byte mutant: whole-tree equality rejected altered core.ts', flush=True)
    finally:
        target.write_bytes(before)


if __name__ == '__main__':
    main()
