#!/usr/bin/env python3
"""Pinned, interleaved best-of-ten user-time measurements with exact output checks."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import time


def observe(command, directory, name, cpu):
    out = directory / (name + ".stdout")
    err = directory / (name + ".stderr")
    before = Path("/proc/loadavg").read_text().strip()
    with out.open("wb") as stdout, err.open("wb") as stderr:
        start = time.perf_counter()
        child = subprocess.Popen(command, stdout=stdout, stderr=stderr,
                                 preexec_fn=lambda: os.sched_setaffinity(0, {cpu}),
                                 env={**os.environ, "GOMAXPROCS": "1"})
        _, status, usage = os.wait4(child.pid, 0)
        elapsed = time.perf_counter() - start
        child.returncode = os.waitstatus_to_exitcode(status)
    if child.returncode or out.read_bytes() != b"0\n" or err.read_bytes():
        raise ValueError(f"output/exit mismatch: {name}")
    return dict(wall=elapsed, user=usage.ru_utime, system=usage.ru_stime,
                command=command, load_before=before,
                load_after=Path("/proc/loadavg").read_text().strip())


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path)
    parser.add_argument("--cpu", type=int, default=3)
    parser.add_argument("--go", type=Path, required=True)
    parser.add_argument("--manifest", type=Path, required=True)
    args = parser.parse_args()
    names = ["baseline", "Go"]
    commands = {n: [str(args.go if n == "Go" else args.directory / n),
                    "--manifest", str(args.manifest), "--count"] for n in names}
    results = {n: [] for n in names}
    for name in names:
        observe(commands[name], args.directory, "warm-" + name, args.cpu)
    for round_number in range(10):
        order = names if round_number % 2 == 0 else list(reversed(names))
        for name in order:
            results[name].append(observe(commands[name], args.directory,
                                         f"round-{round_number + 1}-{name}", args.cpu))
    best = {name: {k: min(row[k] for row in rows) for k in ("user", "wall", "system")}
            for name, rows in results.items()}
    (args.directory / "timing.json").write_text(json.dumps(dict(cpu=args.cpu,
                    samples=results, best=best), indent=2) + "\n")
    print(json.dumps(best, indent=2))


if __name__ == "__main__":
    main()
