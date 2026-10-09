#!/usr/bin/env python3
"""Measure fixture shards and prove their disjoint union and failure isolation.

Pass a prepared fixtures test binary. Builds are deliberately outside measured
units. Every invocation uses a new process and uncached oracle observations.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time


def assignment(name, count):
    return int.from_bytes(hashlib.sha256(name.encode()).digest()[:8], 'big') % count


def audit(binary, output, count):
    root = Path(__file__).resolve().parent
    output.mkdir(parents=True, exist_ok=False)
    cases = {}
    for status in sorted(root.glob('*/status.json')):
        for entry in json.loads(status.read_text()):
            name = status.parent.name + '/' + entry['file']
            if name in cases:
                raise AssertionError('duplicate fixture ' + name)
            cases[name] = assignment(name, count)
    if not cases:
        raise AssertionError('no fixture cases')
    (output / 'cases.json').write_text(json.dumps(cases, indent=2) + '\n')
    rows = []

    def run(label, selector, shard, fixture_root=root):
        log = output / (label + '.log')
        command = [str(binary), '-test.run=' + selector, '-test.count=1',
                   '-test.parallel=4', '-test.timeout=30m', '-test.v',
                   '-fixtures', str(fixture_root)]
        environment = dict(os.environ, GOMAXPROCS='4', ADAMIC_GATE_UNCACHED='1',
                           ADAMIC_TEST_SHARD=shard)
        begin = time.monotonic()
        with log.open('w') as stream:
            result = subprocess.run(command, cwd=root, env=environment,
                                    stdout=stream, stderr=subprocess.STDOUT)
        row = dict(unit=label, seconds=round(time.monotonic() - begin, 3),
                   exit=result.returncode, command=command, shard=shard, log=str(log))
        rows.append(row)
        (output / 'results.json').write_text(json.dumps(rows, indent=2) + '\n')
        print(label, row['seconds'], result.returncode, flush=True)
        return row, log.read_text()

    manifest, text = run('manifest', '^TestFixtureShardManifest$', f'0/{count}')
    observed = {name: int(index) for name, index, total in
                re.findall(r'\s(\S+) shard=(\d+)/(\d+)', text)}
    assert manifest['exit'] == 0 and observed == cases, 'Go manifest disagrees with independent census'
    union = set()
    for index in range(count):
        row, text = run(f'shard-{index}', '^TestFixtures$', f'{index}/{count}')
        assert row['exit'] == 0, row
        names = re.findall(r'^=== RUN\s+TestFixtures/(\S+)\s*$', text, re.M)
        selected = [name for name in names if name in cases]
        expected = {name for name, shard in cases.items() if shard == index}
        assert len(selected) == len(set(selected)), 'duplicate case execution'
        assert set(selected) == expected, (index, selected, expected)
        assert not union.intersection(selected), 'overlapping shards'
        union.update(selected)
        assert row['seconds'] < 30, row
    assert union == set(cases), 'missing cases'

    # Mutate an actual recorded observation, leaving source and other records
    # untouched. Exactly the owning shard must catch its recorded-Node mismatch.
    mutant = output / 'mutant-fixtures'
    shutil.copytree(root, mutant, ignore=shutil.ignore_patterns('__pycache__', 'shard-evidence'))
    status = mutant / 'runner/status.json'
    entries = json.loads(status.read_text())
    target = 'runner/' + entries[-1]['file']
    entries[-1]['node']['stdout'] += 'planted shard failure\n'
    status.write_text(json.dumps(entries, indent=2) + '\n')
    failures = []
    for index in range(count):
        row, text = run(f'mutant-{index}', '^TestFixtures$', f'{index}/{count}', mutant)
        if row['exit']:
            assert 'recorded Node byte comparison failed' in text, text
            assert '--- FAIL: TestFixtures/' + target + '/node' in text, text
            failures.append(index)
        assert row['seconds'] < 30, row
    assert failures == [cases[target]], (failures, cases[target])

    for number, invalid in enumerate(['bad', '0/0', '-1/8', '8/8', '0/2/3']):
        row, text = run(f'invalid-{number}', '^TestFixtureShardManifest$', invalid)
        assert row['exit'] != 0 and 'invalid ADAMIC_TEST_SHARD' in text, text
    for name in ['TestFixturePaths', 'TestTransformedNodeRunnerGuard', 'TestTransformedNodeRunnerGuardHook']:
        row, text = run(name, '^' + name + '$', '')
        assert row['exit'] == 0 and row['seconds'] < 30, row
    proof = dict(fixtures=len(cases), shards=count, union='complete and disjoint',
                 planted_failure=target, caught_shards=failures,
                 binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest())
    (output / 'proof.json').write_text(json.dumps(proof, indent=2) + '\n')
    print(json.dumps(proof, indent=2))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('--count', type=int, default=8)
    args = parser.parse_args()
    if args.count < 1:
        parser.error('count must be positive')
    audit(args.binary.resolve(), args.output.resolve(), args.count)
