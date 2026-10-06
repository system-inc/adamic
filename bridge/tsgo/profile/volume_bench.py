#!/usr/bin/env python3
"""Alternate complete count-only runs. Preserve outputs and separate load/query/run."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import time


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("native")
    parser.add_argument("oracle")
    parser.add_argument("directory", type=Path)
    parser.add_argument("--corpus", nargs=3, action="append", required=True,
                        metavar=("NAME", "CONFIG", "MANIFEST"))
    parser.add_argument("--rounds", type=int, default=3)
    parser.add_argument("--before", help="Optional unchanged native baseline binary")
    args = parser.parse_args()
    args.directory.mkdir(parents=True, exist_ok=True)
    records = []
    environment = dict(os.environ, ADAMIC_TSGO_TIMING="1")
    for name, config, manifest in args.corpus:
        for round_number in range(1, args.rounds + 1):
            order = [("go", args.oracle), ("native", args.native)]
            if args.before:
                order.append(("before", args.before))
            if round_number % 2 == 0:
                order.reverse()
            expected = None
            for implementation, binary in order:
                stem = args.directory / f"{name}-{round_number}-{implementation}"
                with stem.with_suffix(".stdout").open("wb") as output, stem.with_suffix(".stderr").open("wb") as report:
                    before = time.perf_counter_ns()
                    result = subprocess.run([binary, config, manifest, "--count"], env=environment,
                                            stdout=output, stderr=report, check=False)
                    elapsed = time.perf_counter_ns() - before
                if result.returncode:
                    raise RuntimeError(f"{stem}: exit {result.returncode}; see saved stderr")
                output = stem.with_suffix(".stdout").read_bytes()
                if expected is not None and expected != output:
                    raise RuntimeError(f"{stem}: count output differs from independent Go")
                expected = output
                count = re.fullmatch(rb"findings (\d+)\n", output)
                if count is None:
                    raise RuntimeError(f"{stem}: unexpected count output")
                record = dict(corpus=name, round=round_number, implementation=implementation,
                              process_ns=elapsed, findings=int(count[1]))
                record.update({key: int(value) for key, value in re.findall(
                    r"(\w+)=(\d+)", stem.with_suffix(".stderr").read_text())})
                records.append(record)
                (args.directory / "results.json").write_text(json.dumps(records, indent=2) + "\n")
                print(f"{name} round {round_number} {implementation}: {elapsed / 1e9:.6f}s; "
                      f"{record['findings']} findings", flush=True)


if __name__ == "__main__":
    main()
