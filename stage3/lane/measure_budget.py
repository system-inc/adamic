#!/usr/bin/env python3
"""Measure saved shard definitions serially or concurrently, with optional CPU contention."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

from shard_plan import LANE, digest, runtime_hash


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('plan', type=Path)
    parser.add_argument('results', type=Path)
    parser.add_argument('--concurrency', type=int, default=1)
    parser.add_argument('--names', nargs='+')
    parser.add_argument('--load', type=int, default=0)
    parser.add_argument('--unit', help=argparse.SUPPRESS)
    args = parser.parse_args()
    plan = json.loads(args.plan.read_text())
    # Saved main definitions use today's verified artifact and runtime, without changing their tasks.
    runtime = runtime_hash()
    for unit in plan['units']:
        definition = {key: value for key, value in unit.items() if key != 'inputs_hash'}
        unit['inputs_hash'] = digest(dict(artifact=plan['artifact'], api=plan['api_digest'],
                                         runtime=runtime, unit=definition))
    plan['runtime'] = runtime
    if args.unit:
        from shard import run_unit
        unit = next(row for row in plan['units'] if row['name'] == args.unit)
        return run_unit(unit['name'], args.results.resolve(), unit['inputs_hash'], plan=plan)
    if args.concurrency < 1 or args.load < 0:
        parser.error('concurrency must be positive and load nonnegative')
    units = plan['units']
    if args.names:
        units = [unit for unit in units if unit['name'] in args.names]
        if set(args.names) != {unit['name'] for unit in units}:
            parser.error('unknown selected unit')
    root = args.results.resolve()
    root.mkdir(parents=True, exist_ok=False)
    (root / 'plan.json').write_text(json.dumps(plan, indent=2) + '\n')
    load = []
    started = time.monotonic()
    try:
        for _ in range(args.load):
            load.append(subprocess.Popen([sys.executable, '-S', '-c', 'while True: pass'],
                        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, start_new_session=True))
        def run(unit):
            command = [sys.executable, '-B', str(Path(__file__).resolve()), str(root / 'plan.json'),
                       str(root), '--unit', unit['name']]
            env = dict(os.environ, NODE_DISABLE_COMPILE_CACHE='1', PYTHONDONTWRITEBYTECODE='1',
                       STAGE3_RESULTS=str(root), STAGE3_CACHE=str(root / 'caches' / unit['name']))
            begin = time.monotonic()
            with (root / (unit['name'] + '.log')).open('w') as log:
                try:
                    code = subprocess.run(command, env=env, stdout=log, stderr=subprocess.STDOUT,
                                          timeout=60).returncode
                except subprocess.TimeoutExpired:
                    code = 124
            path = root / unit['name'] / 'report.json'
            report = json.loads(path.read_text()) if path.exists() else {}
            row = dict(name=unit['name'], seconds=time.monotonic() - begin, exit=code,
                       status=report.get('status', 'missing'), counts=report.get('counts'),
                       errors=report.get('errors'), command=command)
            return row
        rows = []
        with ThreadPoolExecutor(max_workers=args.concurrency) as pool:
            for row in pool.map(run, units):
                rows.append(row)
                with (root / 'ledger.jsonl').open('a') as stream:
                    stream.write(json.dumps(row) + '\n')
                print(json.dumps(row), flush=True)
        report = dict(concurrency=args.concurrency, load_processes=args.load,
                      mode='adapted-tree product' if os.environ.get('STAGE3_ADAPTED_TREE') else 'shared artifact',
                      adapted_tree=os.environ.get('STAGE3_ADAPTED_TREE'),
                      shard_limit_seconds=float(os.environ.get('STAGE3_SHARD_SECONDS', '30')),
                      affinity=sorted(os.sched_getaffinity(0)),
                      cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
                      cold='fresh processes; NODE_DISABLE_COMPILE_CACHE=1; no result reuse; OS page cache retained',
                      wall_seconds=time.monotonic() - started, units=rows)
        (root / 'ledger.json').write_text(json.dumps(report, indent=2) + '\n')
        return int(any(row['exit'] for row in rows))
    finally:
        for child in load:
            os.killpg(child.pid, signal.SIGKILL)
            child.wait()


if __name__ == '__main__':
    raise SystemExit(main())
