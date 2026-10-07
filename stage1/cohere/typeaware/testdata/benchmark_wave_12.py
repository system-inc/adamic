"""Three interleaved whole-process runs; every round compares full bytes."""
import hashlib
import json
import os
from pathlib import Path
import re
import statistics
import subprocess
import sys
import time

native, oracle, destination, compiler_config, compiler_roots, repository_config, repository_roots = sys.argv[1:]
destination = Path(destination)
destination.mkdir(parents=True, exist_ok=True)
rows = []
for corpus, config, roots in [("compiler", compiler_config, compiler_roots), ("repository", repository_config, repository_roots)]:
    for round_number in range(3):
        outputs = {}
        for implementation in (["native", "go"] if round_number % 2 == 0 else ["go", "native"]):
            stem = destination / f"{corpus}-{round_number}-{implementation}"
            environment = dict(os.environ)
            if implementation == "native":
                environment["ADAMIC_TSGO_TIMING"] = "1"
            with stem.with_suffix(".stdout").open("wb") as stdout, stem.with_suffix(".stderr").open("wb") as stderr:
                started = time.perf_counter()
                result = subprocess.run([native if implementation == "native" else oracle, config, roots], env=environment, stdout=stdout, stderr=stderr)
                elapsed = time.perf_counter() - started
            if result.returncode:
                raise RuntimeError(f"{stem}: exit {result.returncode}")
            output = stem.with_suffix(".stdout").read_bytes()
            phases = {key: int(value) for key, value in re.findall(r"(\w+)=(\d+)", stem.with_suffix(".stderr").read_text())}
            outputs[implementation] = output
            rows.append(dict(corpus=corpus, round=round_number, implementation=implementation, elapsed_seconds=elapsed, phases=phases, bytes=len(output), sha256=hashlib.sha256(output).hexdigest()))
        if outputs["native"] != outputs["go"]:
            raise RuntimeError(f"{corpus} round {round_number}: diagnostic bytes differ")
summary = []
for corpus in ["compiler", "repository"]:
    for implementation in ["native", "go"]:
        selected = [row for row in rows if row["corpus"] == corpus and row["implementation"] == implementation]
        summary.append(dict(corpus=corpus, implementation=implementation, whole_process_median_seconds=statistics.median(row["elapsed_seconds"] for row in selected), load_median_seconds=statistics.median(row["phases"]["load_ns"] / 1e9 for row in selected), run_median_seconds=statistics.median(row["phases"]["run_ns"] / 1e9 for row in selected)))
(destination / "measurements.json").write_text(json.dumps(dict(rounds=rows, summary=summary), indent=2) + "\n")
print(json.dumps(summary, indent=2))
