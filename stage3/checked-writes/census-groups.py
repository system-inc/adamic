#!/usr/bin/env python3
"""Compare relation admission before this unit, after FlowNode, and after containers."""
import argparse
import json
import os
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('adapted')
parser.add_argument('records')
parser.add_argument('output')
parser.add_argument('--baseline', default='b94f41ee')
args = parser.parse_args()
root = Path(__file__).resolve().parents[2]
output = Path(args.output).resolve()
output.mkdir(parents=True, exist_ok=True)
for mode in ['baseline', 'flow-only', 'containers']:
    folder = output / mode
    folder.mkdir(exist_ok=True)
    with (folder / 'setup.log').open('w') as log:
        subprocess.run(['python3', str(root / 'stage3/census/latent/make_overlay.py'), str(root), str(folder)], cwd=root, stdout=log, stderr=subprocess.STDOUT, check=True)
    path = folder / 'overlay.json'
    overlay = json.loads(path.read_text())
    overlay['Replace'][str(root / 'internal/lower/checked_writes_census_test.go')] = str(root / 'stage3/checked-writes/census_test.go.txt')
    if mode == 'baseline':
        # Only these relation routines change admission. Keep the latent loader and
        # output guards intact; allocation emission is never run by this census.
        for name in ['checked_writes.go', 'reference_contracts.go']:
            target = folder / name
            target.write_bytes(subprocess.check_output(['git', 'show', args.baseline + ':internal/lower/' + name], cwd=root))
            overlay['Replace'][str(root / 'internal/lower' / name)] = str(target)
    elif mode == 'flow-only':
        source = (root / 'internal/lower/checked_writes.go').read_text()
        start = source.index('\tif own, view := l.containerRelation(')
        end = source.index('\tif len(l.containers(from))', start)
        target = folder / 'checked_writes.go'
        target.write_text(source[:start] + source[end:])
        overlay['Replace'][str(root / 'internal/lower/checked_writes.go')] = str(target)
    path.write_text(json.dumps(overlay))
    env = dict(os.environ, CHECKED_WRITES_TREE=str(Path(args.adapted).resolve()), CHECKED_WRITES_RECORDS=str(Path(args.records).resolve()), CHECKED_WRITES_OUTPUT=str(folder / 'census.json'))
    with (folder / 'census.log').open('w') as log:
        subprocess.run(['go', 'test', '-overlay', str(path), './internal/lower', '-run', '^TestCheckedWritesCensus$', '-count=1', '-v', '-timeout', '10m'], cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT, check=True)
    print(mode, json.loads((folder / 'census.json').read_text())['class_d_counts'], flush=True)
