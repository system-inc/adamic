#!/usr/bin/env python3
"""Optional clang timing wrapper; ADAMIC_REAL_CLANG must name the real compiler."""
import json
import os
import pathlib
import signal
import subprocess
import sys
import time

compiler = os.environ["ADAMIC_REAL_CLANG"]
started = time.monotonic_ns()
result = subprocess.run([compiler, *sys.argv[1:]])
elapsed = time.monotonic_ns() - started
directory = os.environ.get("ADAMIC_CLANG_MEASURE")
if directory and "--version" not in sys.argv:
    path = pathlib.Path(directory) / f"{started}-{os.getpid()}.json"
    path.write_text(json.dumps({"start_ns": started, "elapsed_ns": elapsed, "args": sys.argv[1:], "exit": result.returncode}))
if result.returncode < 0:
    if -result.returncode not in (signal.SIGKILL, signal.SIGSTOP):
        signal.signal(-result.returncode, signal.SIG_DFL)
    os.kill(os.getpid(), -result.returncode)
sys.exit(result.returncode)
