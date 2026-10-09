#!/usr/bin/env python3
"""Run merged-loader checking for each adapted compiler root and the whole tree."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import gzip
import json
import os
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--probe', type=Path, required=True)
parser.add_argument('--tree', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
roots = sorted(str(p) for p in (args.tree / 'src/compiler').rglob('*.ts'))
args.output.mkdir(parents=True, exist_ok=True)
environment = dict(os.environ, LEDGER_TREE=str(args.tree), LEDGER_STRICT='1')
for mode in ['project', 'strict', 'census']:
    env = dict(environment, LEDGER_STRICT='0' if mode == 'project' else '1', LEDGER_PROFILE='census' if mode == 'census' else 'project')
    result = subprocess.run([str(args.probe), '--whole', *roots], env=env, check=True, text=True, capture_output=True)
    (args.output / ('stage0-' + mode + '-whole.jsonl')).write_text(result.stdout)
    observation = json.loads(result.stdout)
    print(mode, 'whole:', len(observation['diagnostics']), 'diagnostics', flush=True)


def one(root):
    result = subprocess.run([str(args.probe), root], env=dict(environment, LEDGER_PROFILE='census'), check=True, text=True, capture_output=True)
    observation = json.loads(result.stdout)
    observation['file'] = str(Path(root).relative_to(args.tree))
    return observation


with ThreadPoolExecutor(max_workers=2) as workers:
    with gzip.open(args.output / 'stage0-per-root.jsonl.gz', 'wt') as output:
        for observation in workers.map(one, roots):
            output.write(json.dumps(observation) + '\n')
            print(observation['file'], len(observation['diagnostics']), flush=True)
print('completed', len(roots), 'individual strict-root census loads')
