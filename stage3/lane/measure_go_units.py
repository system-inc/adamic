#!/usr/bin/env python3
"""Measure each lane Go unit in a fresh bounded process with prepared products."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('results', type=Path)
    parser.add_argument('--only', default='^Test')
    parser.add_argument('--resume', action='store_true')
    args = parser.parse_args()
    base = args.results.resolve()
    base.mkdir(parents=True, exist_ok=args.resume)
    for name in ['logs', 'shards', 'python']:
        (base / name).mkdir(exist_ok=args.resume)
    list_command = ['go', 'test', './stage3/lane', '-list', '^Test', '-timeout', '90s']
    with (base / 'list.log').open('w') as out:
        result = subprocess.run(list_command, cwd=ROOT, stdout=out, stderr=subprocess.STDOUT, timeout=90)
    if result.returncode:
        raise RuntimeError('Go test discovery failed')
    names = [name for name in (base / 'list.log').read_text().splitlines()
             if name.startswith('Test') and re.search(args.only, name)]
    if not names:
        raise RuntimeError('empty Go unit inventory')
    ledger = base / 'ledger.jsonl'
    passed = {row['name'] for row in map(json.loads, ledger.read_text().splitlines()) if row['exit'] == 0} if ledger.exists() else set()
    env = dict(os.environ, STAGE3_UNION_RESULTS=str(base / 'shards'),
               STAGE3_PYTHON_RESULTS=str(base / 'python'), PYTHONDONTWRITEBYTECODE='1')
    started = time.monotonic()
    for index, name in enumerate(names):
        if name in passed:
            continue
        command = ['go', 'test', './stage3/lane', '-run', '^' + name + '$', '-count=1', '-timeout', '90s', '-json']
        (base / 'current.json').write_text(json.dumps(dict(index=index, total=len(names), name=name, started=time.time())) + '\n')
        log = base / 'logs' / (name + '.jsonl')
        begin = time.monotonic()
        with log.open('w') as out:
            process = subprocess.Popen(command, cwd=ROOT, env=env, stdout=out, stderr=subprocess.STDOUT, start_new_session=True)
            try:
                code = process.wait(timeout=90)
            except subprocess.TimeoutExpired:
                import signal
                os.killpg(process.pid, signal.SIGKILL)
                process.wait(); code = 124
        seconds = time.monotonic() - begin
        events = [json.loads(line) for line in log.read_text().splitlines() if line.startswith('{')]
        finished = [row for row in events if row.get('Test') == name and row.get('Action') in ('pass','fail','skip')]
        wall = finished[-1]['Elapsed'] if finished else None
        if not finished or finished[-1]['Action'] != 'pass':
            code = code or 1
        row = dict(name=name, command=command, exit=code, test_seconds=wall,
                   process_seconds=seconds, flagged_over_30=wall is not None and wall > 30,
                   failure_over_60=wall is not None and wall > 60)
        with ledger.open('a') as out:out.write(json.dumps(row) + '\n')
        print(json.dumps(row), flush=True)
        if code or row['failure_over_60']:
            raise SystemExit(1)
    rows = [json.loads(line) for line in ledger.read_text().splitlines()]
    (base / 'measurements.json').write_text(json.dumps(dict(
        wall_seconds=time.monotonic()-started, nproc=os.cpu_count(),
        cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
        cold='one fresh go test process and fresh outputs per unit; products prepared; OS page cache retained',
        attempts=rows, units=sorted({row['name']: row for row in rows}.values(),key=lambda row:row['test_seconds'] or 0,reverse=True)),indent=2)+'\n')


if __name__ == '__main__':
    main()
