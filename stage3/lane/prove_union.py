#!/usr/bin/env python3
"""Run the real same-artifact union and drop one captured shard to prove rejection."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('measurements', type=Path)
    parser.add_argument('whole', type=Path)
    parser.add_argument('artifact', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    env = dict(os.environ, STAGE3_ARTIFACT=str(args.artifact), PYTHONDONTWRITEBYTECODE='1')
    ledger = []
    def compare(label, wanted):
        report = args.output / (label + '.json')
        command = ['python3', '-B', 'stage3/lane/merge_shards.py', str(args.measurements / 'shards'),
                   '--whole', str(args.whole), '--output', str(report)]
        start = time.monotonic()
        with (args.output / (label + '.log')).open('w') as output:
            result = subprocess.run(command, cwd=ROOT, env=env, stdout=output, stderr=subprocess.STDOUT, timeout=90)
        ledger.append(dict(command=command, exit=result.returncode, seconds=time.monotonic()-start))
        if result.returncode != wanted:
            raise ValueError(f'{label}: expected exit {wanted}, got {result.returncode}')
        return json.loads(report.read_text())
    positive = compare('merged', 0)
    if positive['units'] != 151 or positive['missing_tests'] or positive['extra_tests']:
        raise ValueError('incomplete positive union')
    manifest = json.loads((ROOT / 'stage3/lane/shards.json').read_text())
    owned = next(row for row in manifest['units'] if row['name'] == 'stage3-lane-file-001')
    expected_missing = sum(owned['expect'].values())
    source = args.measurements / 'shards/stage3-lane-file-001'
    removed = args.output / 'removed-shard'
    source.rename(removed)
    try:
        negative = compare('drop-shard', 1)
        if not any('missing shard stage3-lane-file-001; missing tests:' in error for error in negative['errors']):
            raise ValueError('missing shard was not named')
        if sum(row['count'] for row in negative['missing_tests']) != expected_missing or negative['extra_tests']:
            raise ValueError('missing test identity inventory was not exact')
    finally:
        removed.rename(source)
    restored = compare('restored', 0)
    if restored != positive:
        raise ValueError('restored union changed')
    (args.output / 'commands.json').write_text(json.dumps(ledger, indent=2)+'\n')
    print(json.dumps(dict(status='pass', tests=positive['tests'], units=positive['units'],
                          counts=positive['counts'], dropped_missing=sum(row['count'] for row in negative['missing_tests'])),indent=2))


if __name__ == '__main__':
    main()
