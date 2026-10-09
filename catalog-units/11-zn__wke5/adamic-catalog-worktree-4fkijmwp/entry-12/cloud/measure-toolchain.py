#!/usr/bin/env python3
"""Measure go.mod's auto download with an older bootstrap, then the warm lookup.

The normal setup trials install a current Go, so auto has nothing to download.
This separate experiment keeps that distinction visible. Each cold lookup has
an empty GOPATH, with a warm lookup immediately afterward on the same checkout.
"""
import json
import os
from pathlib import Path
import platform
import re
import subprocess
import tempfile
import time

repository = Path(__file__).resolve().parent.parent
scratch = Path(tempfile.mkdtemp(prefix="setup-toolchain-", dir="/workspace"))
bootstrap = "go1.26.0"
architecture = {"x86_64": "amd64", "aarch64": "arm64"}[platform.machine()]
archive = scratch / "bootstrap.tar.gz"
with (scratch / "bootstrap.log").open("wb") as log:
    subprocess.run(["curl", "-fsSL", f"https://dl.google.com/go/{bootstrap}.linux-{architecture}.tar.gz",
                    "-o", str(archive)], stdout=log, stderr=log, check=True)
    subprocess.run(["tar", "--no-same-owner", "-xz", "-C", str(scratch), "-f", str(archive)],
                   stdout=log, stderr=log, check=True)
results = []
for loop in range(1, 4):
    environment = dict(os.environ, GOTOOLCHAIN="auto", GOPATH=str(scratch / f"modules-{loop}"))
    for mode in ["cold", "warm"]:
        before = Path("/proc/loadavg").read_text().strip()
        command = [str(scratch / "go/bin/go"), "version"]
        log_path = scratch / f"{mode}-{loop}.log"
        started = time.monotonic()
        with log_path.open("wb") as log:
            result = subprocess.run(command, cwd=repository, env=environment, stdout=log, stderr=log)
        elapsed = time.monotonic() - started
        after = Path("/proc/loadavg").read_text().strip()
        answer = log_path.read_text()
        # The Go proxy's denied redirect includes a temporary signed URL. Preserve the
        # destination and failure without publishing its query string in an evidence log.
        answer = re.sub(r'https://storage.googleapis.com/[^"\s]+',
                        'https://storage.googleapis.com/<redirect-redacted>', answer)
        log_path.write_text(answer)
        entry = dict(loop=loop, mode=mode, seconds=elapsed, instrument=command,
                     commit=subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repository).decode().strip(),
                     nproc=subprocess.check_output(["nproc"]).decode().strip(),
                     cpu_max=Path("/sys/fs/cgroup/cpu.max").read_text().strip(),
                     go=answer.strip(), bootstrap=bootstrap,
                     clang=subprocess.check_output(["clang", "--version"]).decode().splitlines()[0],
                     node=subprocess.check_output(["node", "--version"]).decode().strip(),
                     load_before=before, load_after=after, cached=(mode == "warm"),
                     exit=result.returncode, log=str(log_path),
                     environment={"GOTOOLCHAIN": "auto", "GOPATH": environment["GOPATH"]})
        results.append(entry)
        (scratch / "timings.json").write_text(json.dumps(results, indent=2) + "\n")
        print(f"{mode} loop={loop} wall={elapsed:.6f}s exit={result.returncode} log={log_path}", flush=True)
        if result.returncode:
            raise SystemExit(result.returncode)
print("results:", scratch / "timings.json")
