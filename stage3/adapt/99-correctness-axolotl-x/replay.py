#!/usr/bin/env python3
"""Replay in apply.py's exact Python directory order; stop at the first refusal or write."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
repo, tree, logdir = map(Path, sys.argv[1:])
logdir.mkdir(parents=True, exist_ok=True)
env = dict(os.environ, NODE_PATH=str(repo / 'stage3/api/node_modules'))
def snapshot():
    return {str(p.relative_to(tree)): (hashlib.sha256(p.read_bytes()).hexdigest(), p.stat().st_mtime_ns)
            for p in tree.rglob('*') if p.is_file() and '.git' not in p.relative_to(tree).parts}
rows = []
for directory in sorted((repo / 'stage3/adapt').iterdir()):
    if not directory.is_dir():
        continue
    before = snapshot()
    with (logdir / (directory.name + '.log')).open('w') as log:
        result = subprocess.run(['node', str(directory / 'adapt.cjs'), str(tree)], env=env, stdout=log, stderr=subprocess.STDOUT)
    after = snapshot()
    changed = [name for name in before.keys() | after.keys() if before.get(name) != after.get(name)]
    details = [dict(file=name, before_sha256=before.get(name, (None, None))[0], after_sha256=after.get(name, (None, None))[0], before_mtime_ns=before.get(name, (None, None))[1], after_mtime_ns=after.get(name, (None, None))[1]) for name in sorted(changed)]
    row = dict(adapter=directory.name, exit=result.returncode, changed=sorted(changed), details=details)
    rows.append(row)
    (logdir / 'results.json').write_text(json.dumps(rows, indent=2) + '\n')
    print(json.dumps(row), flush=True)
    if result.returncode or changed:
        sys.exit(1)
