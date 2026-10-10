#!/usr/bin/env python3
"""Measure selected oracle tests with cold runtime caches and hard deadlines.

Go compilation is provisioning and must be measured separately. Run the binary
from internal/oracle, as go test does. Use --run repeatedly to rerun only changed
or failed shards. A cache directory that already exists is rejected.
"""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument("--binary", type=Path, required=True)
parser.add_argument("--out", type=Path, required=True)
parser.add_argument("--run", action="append", required=True)
args = parser.parse_args()
def cancel(signum, frame):
    raise KeyboardInterrupt
signal.signal(signal.SIGTERM, cancel)
binary = args.binary.resolve()
root = Path(__file__).resolve().parents[2]
args.out.mkdir(parents=True, exist_ok=True)
rows = []
for name in args.run:
    label = name.replace("/", "-")
    cache = args.out.resolve() / (label + "-cache")
    if cache.exists():
        raise SystemExit(f"cold cache already exists: {cache}")
    env = dict(os.environ, GOMAXPROCS="4", ADAMIC_GATE_UNCACHED="1", XDG_CACHE_HOME=str(cache))
    start = time.monotonic()
    killed = False
    with (args.out / (label + ".log")).open("w") as log:
        process = subprocess.Popen(
            [str(binary), "-test.run", "^" + name.replace("/", "$/^") + "$",
             "-test.count=1", "-test.parallel=4", "-test.timeout=90s", "-test.v"],
            cwd=root, env=env, stdout=log, stderr=subprocess.STDOUT,
            start_new_session=True)
        try:
            code = process.wait(timeout=90)
        except subprocess.TimeoutExpired:
            killed = True
            os.killpg(process.pid, signal.SIGKILL)
            code = process.wait()
        finally:
            if process.poll() is None:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
    wall = time.monotonic() - start
    row = dict(test=name, wall_seconds=round(wall, 3), killed_at_deadline=killed,
               exit_code=code, assertions_passed=code == 0,
               gate_passed=code == 0 and wall < 60, under_45_seconds=wall < 45)
    rows.append(row)
    (args.out / "summary.json").write_text(json.dumps(rows, indent=2) + "\n")
    print(json.dumps(row), flush=True)
    if code != 0 or wall >= 45:
        raise SystemExit("FAIL: split or fix this unit, then rerun only its replacements")
