#!/usr/bin/env python3
"""Measure identical uncached oracle work with clang and GCC, alternating order."""
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import time


def output(*arguments):
    return subprocess.check_output(arguments, text=True).strip()


metadata = {
    "commit": output("git", "rev-parse", "HEAD"),
    "nproc": output("nproc"),
    "cpu.max": Path("/sys/fs/cgroup/cpu.max").read_text().strip(),
    "go": output("go", "version"),
    "clang": output("clang", "--version").splitlines()[0],
    "gcc": output("gcc", "--version").splitlines()[0],
    "node": output("node", "--version"),
    "cache": "uncached observations; fresh runtime and fixture binaries; Go action cache enabled",
}
rows = []
for mode in ["release", "sanitize"]:
    for round_number, order in enumerate([("clang", "gcc"), ("gcc", "clang"), ("clang", "gcc")], 1):
        for compiler in order:
            name = f"final-{mode}-{round_number}-{compiler}"
            log = Path(f"/tmp/gcc-{name}.log")
            settings = {
                "ADAMIC_GATE_UNCACHED": "1",
                "ADAMIC_GCC_LANE": "1",
                "ADAMIC_LANE_CC": compiler,
                "ADAMIC_LANE_SANITIZE": str(int(mode == "sanitize")),
                "ADAMIC_LANE_REPORT": f"/tmp/gcc-{name}",
            }
            command = ["go", "test", "./internal/oracle", "-run", "^TestGCCAgreesWithNode$", "-v", "-count=1", "-timeout", "30m"]
            load_before = Path("/proc/loadavg").read_text().strip()
            with log.open("w") as stream:
                started = time.monotonic()
                result = subprocess.run(command, env=dict(os.environ, **settings), stdout=stream, stderr=subprocess.STDOUT)
                elapsed = time.monotonic() - started
            load_after = Path("/proc/loadavg").read_text().strip()
            flags = re.search(r"compiler=\S+ flags=(.*)", log.read_text())[1]
            rows.append(dict(metadata, mode=mode, round=round_number, compiler=compiler, seconds=elapsed,
                             exit=result.returncode, flags=flags, load_before=load_before, load_after=load_after,
                             instrument="Python time.monotonic around subprocess.run",
                             command=" ".join(f"{key}={shlex.quote(value)}" for key, value in settings.items()) + " " + shlex.join(command) + f" > {log} 2>&1"))
            Path("/tmp/gcc-lane-completion-timings.json").write_text(json.dumps(rows, indent=2) + "\n")
            print(name, result.returncode, f"{elapsed:.6f}", flush=True)
