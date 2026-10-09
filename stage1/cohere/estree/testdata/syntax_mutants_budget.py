#!/usr/bin/env python3
"""Run each TestSyntaxMutants shard with a 60 s budget and a hard 75 s kill.

Build products should already be available by hash. Cooked is a budget result,
separate from a semantic failure. Killing the process group also stops builders
and oracle children. The count comes from the fixed Go planner declaration.
"""
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import time

root = Path(__file__).resolve().parents[4]
source = (root / "stage1/cohere/estree/syntax_test.go").read_text()
count = int(re.search(r"const testSyntaxMutantsShards = (\d+)", source)[1])
results = []
for shard in range(count):
    name = f"TestSyntaxMutants_{shard:03d}"
    environment = dict(os.environ, ADAMIC_TEST_SHARD=f"{shard}/{count}")
    started = time.monotonic()
    process = subprocess.Popen(
        ["go", "test", "-json", "./stage1/cohere/estree", "-run",
         "^" + name + "$", "-count=1", "-timeout", "75s"],
        cwd=root, env=environment, start_new_session=True,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
    )
    killed = False
    try:
        stdout, stderr = process.communicate(timeout=75)
    except subprocess.TimeoutExpired:
        killed = True
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        stdout, stderr = process.communicate()
    wall = time.monotonic() - started
    events = []
    for line in stdout.splitlines():
        if line.startswith("{"):
            try:
                events.append(json.loads(line))
            except json.JSONDecodeError:
                pass  # A hard kill may truncate the final event.
    passed = [event for event in events if event.get("Test") == name and event["Action"] == "pass"]
    cooked = killed or wall > 60 or "test timed out after" in stdout + stderr
    elapsed = passed[0]["Elapsed"] if passed else None
    status = "cooked" if cooked else "pass" if process.returncode == 0 and len(passed) == 1 else "fail"
    results.append(status)
    print(json.dumps({"shard": name, "elapsed": elapsed, "wall": round(wall, 3),
                      "cooked": cooked, "status": status}), flush=True)
    if status == "fail":
        print(stdout + stderr, flush=True)
    if cooked:
        break  # Split the cooked unit before rerunning.
raise SystemExit(124 if "cooked" in results else 1 if "fail" in results else 0)
