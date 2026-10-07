"""Quiet alternating whole-process measurements with full-stream equality."""
import hashlib
import json
import os
from pathlib import Path
import re
import statistics
import subprocess
import sys
import time

artifacts = Path(sys.argv[1])
output = Path(sys.argv[2])
repository = Path(__file__).resolve().parents[4]
compiler = Path(sys.argv[3])
output.mkdir(parents=True, exist_ok=True)
environment = dict(os.environ, ADAMIC_TSGO_TIMING="1")
rows = []
for name, config in [("compiler", compiler / "src/compiler/tsconfig.json"), ("repository", repository / "tsconfig.json")]:
    manifest = artifacts / (name + ".manifest")
    for round_ in range(1, 4):
        subjects = ["go", "native"] if round_ % 2 else ["native", "go"]
        hashes = []
        for subject in subjects:
            binary = artifacts / ("wave17-third-oracle" if subject == "go" else "wave17-third")
            stem = output / f"{name}-{round_}-{subject}"
            with stem.with_suffix(".stdout").open("wb") as stdout, stem.with_suffix(".stderr").open("wb") as stderr:
                started = time.perf_counter()
                subprocess.run([str(binary), str(config), str(manifest)], env=environment, stdout=stdout, stderr=stderr, check=True)
                elapsed = time.perf_counter() - started
            data = stem.with_suffix(".stdout").read_bytes()
            digest = hashlib.sha256(data).hexdigest()
            hashes.append(digest)
            fields = {k: int(v) for k, v in re.findall(r"(\w+)=(\d+)", stem.with_suffix(".stderr").read_text())}
            rows.append(dict(corpus=name, round=round_, subject=subject, process_seconds=elapsed, bytes=len(data), sha256=digest, **fields))
        if hashes[0] != hashes[1]:
            raise RuntimeError("timed diagnostic bytes differ")
medians = []
for name in ["compiler", "repository"]:
    for subject in ["native", "go"]:
        selected = [r for r in rows if r["corpus"] == name and r["subject"] == subject]
        medians.append(dict(corpus=name, subject=subject, process_seconds=statistics.median(r["process_seconds"] for r in selected), load_seconds=statistics.median(r["load_ns"] for r in selected) / 1e9, run_seconds=statistics.median(r["run_ns"] for r in selected) / 1e9, queries=selected[0].get("queries")))
(output / "measurements.json").write_text(json.dumps(dict(rounds=rows, medians=medians), indent=2) + "\n")
print(json.dumps(medians, indent=2))
