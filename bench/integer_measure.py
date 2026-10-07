#!/usr/bin/env python3
"""Interleave release binaries from two compiler builds with the same source on Node."""
import argparse
import datetime
import json
import os
from pathlib import Path
import statistics
import subprocess
import tempfile
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--before", type=Path, required=True)
parser.add_argument("--after", type=Path, required=True)
parser.add_argument("--rounds", type=int, default=3)
parser.add_argument("--regressions", action="store_true")
args = parser.parse_args()
repository = Path(__file__).resolve().parents[1]
compilers = {"before": args.before.resolve(), "after": args.after.resolve()}
sources = [repository / "bench" / name for name in ("integer_bitwise.a", "integer_remainder.a")]
if args.regressions:
    sources.extend(sorted((repository / "bench").glob("*.ts")))
metadata = {
    "date": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "nproc": subprocess.check_output(["nproc"], text=True).strip(),
    "cpu_max": Path("/sys/fs/cgroup/cpu.max").read_text().strip(),
    "machine": os.uname().machine,
    "node": subprocess.check_output(["node", "--version"], text=True).strip(),
    "clang": subprocess.check_output(["clang", "--version"], text=True).splitlines()[0],
    "load_before": os.getloadavg(),
    "native_flags": "adamic build: clang -O2, -ffp-contract=off, no sanitizers",
}
rows = []
with tempfile.TemporaryDirectory(prefix="adamic-integers-") as directory:
    commands = {}
    for source in sources:
        commands[source.stem] = {"node": ["node", str(source)]}
        for version, compiler in compilers.items():
            binary = Path(directory) / (source.stem + "-" + version)
            subprocess.run([str(compiler), "build", str(source), "-o", str(binary)], check=True)
            commands[source.stem][version] = [str(binary)]
    for round_index in range(args.rounds):
        for source in sources:
            answers = []
            for version in ("before", "after", "node"):
                start = time.perf_counter()
                result = subprocess.run(commands[source.stem][version], capture_output=True, text=True, check=True, timeout=180)
                seconds = time.perf_counter() - start
                if result.stderr:
                    raise RuntimeError(result.stderr)
                answers.append(result.stdout)
                row = {"round": round_index + 1, "case": source.stem, "backend": version, "seconds": seconds, "stdout": result.stdout.strip()}
                rows.append(row)
                print(json.dumps(row), flush=True)
            if len(set(answers)) != 1:
                raise RuntimeError(f"{source.name}: outputs differ")
metadata["load_after"] = os.getloadavg()
medians = {}
for source in sources:
    medians[source.stem] = {
        version: statistics.median(row["seconds"] for row in rows if row["case"] == source.stem and row["backend"] == version)
        for version in ("before", "after", "node")
    }
print(json.dumps({"metadata": metadata, "medians": medians, "runs": rows}, indent=2))
