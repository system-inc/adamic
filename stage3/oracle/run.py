#!/usr/bin/env python3
"""Run upstream's locked build and tests; preserve logs and baseline differences."""
import argparse
import difflib
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import time
from failure_details import read_failures, cause_counts, diff_previews, markdown

parser = argparse.ArgumentParser()
parser.add_argument('tree', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('--runners', default='all', help='all, or upstream comma-separated runner names')
parser.add_argument('--tests', help='upstream test regex, for targeted mutant checks')
# Eight workers, fixed: the verdicts are identical at 4, 8, 16 and 24 (106,366 tests, Oct 7), and 8
# ran the suite in 134 s against 225 s at 4 and 135 s at 24 on a 24-CPU slot. One per CPU went red on
# the whole gate, where 64 workers beside every other package tripped upstream's 40 s hook timeout.
parser.add_argument('--workers', type=int, default=8)
parser.add_argument('--limit-seconds', type=int, default=2400)
parser.add_argument('--timeout', type=int, help='upstream Mocha timeout in ms, targeted probes only')
parser.add_argument('--prepared', action='store_true', help='internal same-artifact lane measurement: reuse producer phases')
args = parser.parse_args()
if args.timeout is not None and (not args.tests or args.timeout <= 0):
    parser.error('--timeout requires --tests and a positive millisecond limit')
# Upstream's parallel host ignores --tests; one worker uses runtests' Mocha grep.
if args.tests:
    args.workers = 1
tree = args.tree.resolve()
out = args.output.resolve()
out.mkdir(parents=True, exist_ok=False)
started = time.monotonic()
report = dict(tree=str(tree), nproc=int(subprocess.check_output(['nproc'])), workers=args.workers,
              runners=args.runners, tests=args.tests, node=subprocess.check_output(['node', '--version'], text=True).strip(),
              phases={}, status='fail')

def phase(name, command, limit=None, env=None):
    begin = time.monotonic()
    with (out / (name + '.log')).open('w') as log:
        child = subprocess.Popen(command, cwd=tree, stdout=log, stderr=subprocess.STDOUT, start_new_session=True, env=env)
        try:
            code = child.wait(timeout=limit)
            # Upstream can exit before its workers when a worker crashes.
            try:
                os.killpg(child.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
        except subprocess.TimeoutExpired:
            os.killpg(child.pid, signal.SIGTERM)
            try:
                child.wait(timeout=10)
            except subprocess.TimeoutExpired:
                os.killpg(child.pid, signal.SIGKILL)
                child.wait()
            code = 124
    report['phases'][name] = dict(command=command, exit=code, seconds=round(time.monotonic()-begin, 3))
    return code

try:
    if args.prepared:
        producer = json.loads((tree.parent / 'artifact-ready.json').read_text())
        for name in ('install', 'build'):
            report['phases'][name] = dict(producer['phases'][name], artifact=producer['key'])
        report['artifact'] = producer['key']
    if args.prepared or phase('install', ['npm', 'ci', '--no-audit', '--no-fund']) == 0:
        if args.prepared or phase('build', ['npm', 'run', 'build']) == 0:
            command = ['npm', 'test', '--', f'--workers={args.workers}', '--lint=false', '--no-colors']
            if args.prepared:
                command = ['node', str(Path(__file__).with_name('run-prepared.mjs')), str(tree),
                           f'--workers={args.workers}', '--light=false', '--no-colors']
            if args.runners != 'all':
                command.append('--runners=' + args.runners)
            if args.tests:
                command.append('--tests=' + args.tests)
            if args.timeout is not None:
                command.append('--timeout=' + str(args.timeout))
                report['timeout_ms'] = args.timeout
            observations = out / 'mocha-errors'
            observations.mkdir()
            env = dict(os.environ, STAGE3_ERROR_DIR=str(observations))
            observer = Path(__file__).with_name('observe-errors.cjs').resolve()
            env['NODE_OPTIONS'] = env.get('NODE_OPTIONS', '') + ' --require=' + json.dumps(str(observer))
            code = phase('tests', command, args.limit_seconds, env)
            log = (out / 'tests.log').read_text(errors='replace')
            report['counts'] = {name: int((re.findall(r'^\s*(\d+) ' + name + r'\b', log, re.MULTILINE) or ['0'])[-1]) for name in ['passing', 'failing', 'pending']}
            # Upstream only writes mismatches into local; absent files were not failures.
            differences = []
            local = tree / 'tests/baselines/local'
            reference = tree / 'tests/baselines/reference'
            with (out / 'baseline.diff').open('w') as diff:
                for file in sorted(local.rglob('*')):
                    if not file.is_file():
                        continue
                    relative = file.relative_to(local)
                    deleted = file.name.endswith('.delete')
                    refname = Path(str(relative)[:-7]) if deleted else relative
                    ref = reference / refname
                    old_bytes = ref.read_bytes() if ref.exists() else b''
                    new_bytes = b'' if deleted else file.read_bytes()
                    if old_bytes != new_bytes or deleted or not ref.exists():
                        old = old_bytes.decode('utf-8', errors='backslashreplace')
                        new = new_bytes.decode('utf-8', errors='backslashreplace')
                        differences.append(str(relative))
                        diff.writelines(difflib.unified_diff(old.splitlines(True), new.splitlines(True), fromfile='reference/' + str(refname), tofile='local/' + str(relative)))
            report['baseline_diffs'] = differences
            report['status'] = 'pass' if code == 0 and not differences and report['counts']['passing'] > 0 else 'fail'
            try:
                diff_text = (out / 'baseline.diff').read_text()
                report['failures'] = read_failures(log, diff_text, observations)
                report['cause_counts'] = cause_counts(report['failures'])
                report['baseline_previews'] = list(diff_previews(diff_text).values())
                (out / 'failures.md').write_text(markdown(report['failures']))
            except (OSError, ValueError, KeyError, TypeError) as error:
                report['failure_details_error'] = str(error)
finally:
    report['wall_seconds'] = round(time.monotonic() - started, 3)
    (out / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))
raise SystemExit(0 if report['status'] == 'pass' else 1)
