#!/usr/bin/env python3
"""Build a private lane tree with the gate product's local/tests tasks."""
import json
import os
from pathlib import Path
import subprocess
import sys
import time

repository, tree, output = map(lambda p: Path(p).resolve(), sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
install = repository / 'stage3/typescript/node_modules'
link = tree / 'node_modules'
if not link.exists():
    link.symlink_to(install, target_is_directory=True)
rows = []
for task in ('local', 'tests'):
    command = ['node', str(install / 'hereby/bin/hereby.js'), task]
    started = time.monotonic()
    with (output / (task + '.log')).open('w') as log:
        result = subprocess.run(command, cwd=tree, env=dict(os.environ, NODE_DISABLE_COMPILE_CACHE='1'), stdout=log, stderr=subprocess.STDOUT)
    row = dict(task=task, command=command, exit=result.returncode, seconds=time.monotonic() - started)
    rows.append(row)
    (output / 'phases.json').write_text(json.dumps(rows, indent=2) + '\n')
    print(json.dumps(row), flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)
# Match the gate product: each shard view supplies its own dependency link.
if link.is_symlink():
    link.unlink()
