#!/usr/bin/env python3
"""Run one bounded piece of the unchanged lane manifest and retain all observations."""
import argparse
import gzip
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import tarfile
import time

parser = argparse.ArgumentParser(description=__doc__)
for name in ('checkout', 'artifact', 'results', 'evidence'):
    parser.add_argument(name, type=Path)
parser.add_argument('piece', type=int)
parser.add_argument('start', type=int)
parser.add_argument('--count', type=int, default=40)
args = parser.parse_args()
checkout, artifact, results, evidence = (getattr(args, key).resolve() for key in ('checkout', 'artifact', 'results', 'evidence'))
results.mkdir(parents=True, exist_ok=True)
evidence.mkdir(parents=True, exist_ok=True)
units = json.loads((checkout / 'stage3/lane/shards.json').read_text())['units']
selected = units[args.start:args.start + args.count]
assert selected and 0 < args.count <= 40
started = time.monotonic()
rows = []

def terminate(child):
    # Include detached test workers before terminating their parent.
    processes = [line.split() for line in subprocess.check_output(['ps', '-eo', 'pid=,ppid='], text=True).splitlines()]
    descendants = {child.pid}
    while True:
        more = {int(pid) for pid, parent in processes if int(parent) in descendants}
        if more <= descendants:
            break
        descendants |= more
    for pid in reversed(sorted(descendants)):
        try:
            if os.getpgid(pid) == pid:
                os.killpg(pid, signal.SIGKILL)
            else:
                os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
    child.wait()

for index, unit in enumerate(selected, args.start):
    remaining = 1380 - (time.monotonic() - started)
    if remaining <= 60:
        break
    env = dict(os.environ, STAGE3_ARTIFACT=str(artifact), STAGE3_RESULTS=str(results),
               STAGE3_CACHE=str(results / 'caches' / unit['name']),
               NODE_DISABLE_COMPILE_CACHE='1', PYTHONDONTWRITEBYTECODE='1')
    env.pop('STAGE3_ADAPTED_TREE', None)
    env.pop('TSC_ADAPT_TYPESCRIPT', None)
    begin = time.monotonic()
    with (results / (unit['name'] + '.log')).open('w') as log:
        child = subprocess.Popen(unit['command'], cwd=checkout, env=env,
                                 stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        try:
            code = child.wait(timeout=min(60, remaining))
        except subprocess.TimeoutExpired:
            terminate(child)
            code = 124
    report_path = results / unit['name'] / 'report.json'
    report = json.loads(report_path.read_text()) if report_path.exists() else {}
    row = dict(index=index, name=unit['name'], command=unit['command'], exit=code,
               seconds=time.monotonic() - begin, status=report.get('status', 'missing'),
               counts=report.get('counts'), errors=report.get('errors'),
               inputs_hash=report.get('inputs_hash'), artifact=report.get('artifact'),
               unit_seconds=report.get('seconds'))
    rows.append(row)
    print(json.dumps(row), flush=True)
    if code or row['status'] != 'pass' or report.get('seconds', 60) >= 30:
        break

archive = evidence / f'piece-{args.piece:02d}.tar.gz'
with archive.open('wb') as raw, gzip.GzipFile(fileobj=raw, mode='wb', mtime=0, filename='') as compressed:
    with tarfile.open(fileobj=compressed, mode='w|') as tar:
        for row in rows:
            unit_root = results / row['name']
            log = results / (row['name'] + '.log')
            tar.add(log, arcname=log.name)
            for directory, directories, files in os.walk(unit_root):
                directories[:] = [name for name in directories if name != 'adapted-tree']
                for name in sorted(files):
                    file = Path(directory) / name
                    if not file.is_symlink():
                        tar.add(file, arcname=str(file.relative_to(results)), recursive=False)
# Reports, observations, diffs and logs remain in results for the official merge.
for row in rows:
    if row['exit'] == 0 and row['status'] == 'pass':
        private_view = results / row['name'] / 'adapted-tree'
        if private_view.exists():
            shutil.rmtree(private_view)
seconds = time.monotonic() - started
passed = len(rows) == len(selected) and all(row['exit'] == 0 and row['status'] == 'pass' and row['unit_seconds'] < 30 for row in rows)
report = dict(piece=args.piece, start=args.start, end=args.start + len(rows),
              checkout_commit=subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=checkout, text=True).strip(),
              artifact=artifact.name, seconds=seconds, under25Minutes=seconds < 1500,
              status='pass' if passed and seconds < 1500 else 'fail',
              manifest_units=len(units), units=rows)
(evidence / f'piece-{args.piece:02d}.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps({key: report[key] for key in ('piece', 'start', 'end', 'seconds', 'under25Minutes', 'status')}), flush=True)
raise SystemExit(0 if report['status'] == 'pass' else 1)
