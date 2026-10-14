#!/usr/bin/env python3
"""Prove rebuilt unrelated inputs pass and stale artifacts/runtime fail independently."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time
from artifact import input_key
from shard_plan import LANE

ROOT = LANE.parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('artifact', type=Path)
    parser.add_argument('results', type=Path)
    args = parser.parse_args()
    out = args.results.resolve(); out.mkdir(parents=True, exist_ok=False)
    manifest = LANE / 'shards.json'
    manifest_bytes = manifest.read_bytes()
    commands = []

    def run(label, command, expected=0, env=None):
        started = time.monotonic()
        log = out / (label + '.log')
        with log.open('w') as stream:
            result = subprocess.run(command, cwd=ROOT, env=env,
                                    stdout=stream, stderr=subprocess.STDOUT, timeout=95)
        row = dict(label=label, command=command, exit=result.returncode,
                   seconds=time.monotonic() - started)
        commands.append(row); print(json.dumps(row), flush=True)
        (out / 'commands.json').write_text(json.dumps(commands, indent=2) + '\n')
        if result.returncode != expected:
            raise AssertionError(f'{label}: exit {result.returncode}, expected {expected}; {log}')
        return log.read_text()

    test = ['go', 'test', './stage3/lane', '-run', '^TestShard_Stage3LaneFile001$',
            '-count=1', '-timeout', '90s', '-v']
    base_key = input_key()[0]
    assert json.loads((args.artifact / 'ready.json').read_text())['key'] == base_key
    census = ROOT / 'stage3/census/tsc-closure/README.md'
    original = census.read_bytes()
    changed_key = None
    try:
        census.write_bytes(original + b'\nRuntime artifact identity proof for Task #hefhych.\n')
        changed_key = input_key()[0]
        assert changed_key != base_key
        for phase in ('Install', 'Compiler', 'Harness'):
            run('rebuilt-' + phase, ['go', 'test', './stage3/lane', '-run',
                '^TestProduct_Stage3Lane' + phase + '$', '-count=1', '-timeout', '90s', '-v'])
        run('rebuilt-input-shard001', test)
        assert manifest.read_bytes() == manifest_bytes
        unit = next(row for row in json.loads(manifest_bytes)['units'] if row['name'] == 'stage3-lane-file-001')
        env = dict(os.environ, STAGE3_ARTIFACT=str(args.artifact.resolve()), STAGE3_RESULTS=str(out / 'stale'))
        text = run('stale-artifact', unit['command'], 1, env)
        assert 'checkout key ' + changed_key in text and 'artifact ready key ' + base_key in text
    finally:
        census.write_bytes(original)
    runtime = LANE / 'shard.py'
    original = runtime.read_bytes()
    try:
        runtime.write_bytes(original + b'\n# Runtime identity mutant for Task #hefhych.\n')
        text = run('changed-runtime', test, 1)
        assert 'shard harness inputs changed' in text
    finally:
        runtime.write_bytes(original)
    assert input_key()[0] == base_key
    assert manifest.read_bytes() == manifest_bytes
    proof = dict(base_key=base_key, changed_key=changed_key,
                 manifest_sha256=hashlib.sha256(manifest_bytes).hexdigest(),
                 manifest_unchanged=True, source_restored=True, commands=commands)
    (out / 'proof.json').write_text(json.dumps(proof, indent=2) + '\n')


if __name__ == '__main__':
    main()
