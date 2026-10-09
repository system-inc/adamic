#!/usr/bin/env python3
"""Measure independent manifest commands on one box; this coordinator is not a test unit."""
import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
import json
import os
from pathlib import Path
import subprocess
import time
from shard_plan import LANE


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('results', type=Path)
    parser.add_argument('--concurrency', type=int, default=2)
    parser.add_argument('--mutant-runner', type=Path)
    parser.add_argument('--only-prefix', help='measure a specified subset for targeted case proofs')
    args = parser.parse_args()
    if args.concurrency < 1:
        parser.error('concurrency must be positive')
    root = args.results.resolve(); root.mkdir(parents=True, exist_ok=False)
    units = json.loads((LANE / 'shards.json').read_text())['units']
    if args.only_prefix:
        units = [unit for unit in units if unit['name'].startswith(args.only_prefix)]
    if not units:
        parser.error('no selected units')
    def run(unit):
        env = dict(os.environ, STAGE3_RESULTS=str(root), STAGE3_CACHE=str(root / 'caches' / unit['name']),
                   NODE_DISABLE_COMPILE_CACHE='1', PYTHONDONTWRITEBYTECODE='1')
        command = list(unit['command'])
        if args.mutant_runner:
            command += ['--mutant-runner', str(args.mutant_runner.resolve())]
        started = time.monotonic()
        with (root / (unit['name'] + '.log')).open('w') as log:
            code = subprocess.run(command, cwd=LANE.parents[1], env=env,
                                  stdout=log, stderr=subprocess.STDOUT).returncode
        seconds = time.monotonic() - started
        report_file = root / unit['name'] / 'report.json'
        report = json.loads(report_file.read_text()) if report_file.exists() else {}
        return dict(name=unit['name'], seconds=seconds, exit=code, status=report.get('status', 'missing'),
                    counts=report.get('counts'), errors=report.get('errors'), inputs_hash=unit['command'][-1])
    started = time.monotonic(); rows = []
    with ThreadPoolExecutor(max_workers=args.concurrency) as pool:
        jobs = [pool.submit(run, unit) for unit in units]
        for job in as_completed(jobs):
            row = job.result(); rows.append(row)
            with (root / 'ledger.jsonl').open('a') as stream:
                stream.write(json.dumps(row) + '\n')
            print(json.dumps(row), flush=True)
    report = dict(concurrency=args.concurrency, nproc=os.cpu_count(),
                  cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                  cold='fresh processes; NODE_DISABLE_COMPILE_CACHE=1; no result reuse; OS page cache retained',
                  wall_seconds=time.monotonic() - started, maximum=max(row['seconds'] for row in rows),
                  units=sorted(rows, key=lambda row: row['name']))
    (root / 'ledger.json').write_text(json.dumps(report, indent=2) + '\n')
    return int(any(row['exit'] or row['seconds'] >= 30 for row in rows))


if __name__ == '__main__':
    raise SystemExit(main())
