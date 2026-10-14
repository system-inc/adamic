#!/usr/bin/env python3
"""Repeat the whole lane and every fresh Go unit before comparing the exact union."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import time

ROOT = Path(__file__).resolve().parents[2]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('artifact', type=Path)
    parser.add_argument('measurements', type=Path, help='resume the requested cold Go runs')
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    out = args.output.resolve(); out.mkdir(parents=True, exist_ok=False)
    rows = []

    def run(label, command, seconds):
        start = time.monotonic()
        with (out / (label + '.log')).open('w') as log:
            child = subprocess.Popen(command, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT,
                                     start_new_session=True)
            try:
                code = child.wait(timeout=seconds)
            except subprocess.TimeoutExpired:
                os.killpg(child.pid, signal.SIGKILL); child.wait(); code = 124
        row = dict(label=label, command=command, exit=code, seconds=time.monotonic() - start)
        rows.append(row); print(json.dumps(row), flush=True)
        (out / 'commands.json').write_text(json.dumps(rows, indent=2) + '\n')
        if code:
            raise SystemExit(code)

    run('whole', ['bash', 'stage3/lane/run.sh', str(out / 'whole'),
                  '--artifact', str(args.artifact.resolve())], 580)
    # Every unit has its own 90-second subprocess bound inside the coordinator.
    # The outer coordinator is explicitly monitored through current.json/ledger.
    run('Go-units', ['python3', '-B', 'stage3/lane/measure_go_units.py',
                    str(args.measurements.resolve()), '--resume'], 7200)
    run('union', ['python3', '-B', 'stage3/lane/prove_union.py',
                 str(args.measurements.resolve()), str(out / 'whole'),
                 str(args.artifact.resolve()), str(out / 'comparison')], 180)
    run('Python-union', ['python3', '-B', 'stage3/lane/prove_python_union.py',
                        str(args.measurements.resolve()), str(out / 'python-comparison')], 180)


if __name__ == '__main__':
    main()
