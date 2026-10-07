#!/usr/bin/env python3
"""Resume the fixed-main reference, calibrate its isolated observer, prove shapes."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import time

from proof import atomic_json, canonical_events, run_package, timing_metadata


def calibration_verdicts(events):
    """The always-selected calibration may print varying benchmark output."""
    return {name: [action for action, _ in entries if action != "output"]
            for name, entries in events.items()}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--record", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--main", required=True)
    parser.add_argument("--binary", default="/tmp/adamic-affected")
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    environment = dict(os.environ, PWD=str(args.root), _=args.binary)
    package = "github.com/system-inc/adamic/stage1/cohere/markdownblocks"
    status_path = args.output / "workflow.json"
    state = json.loads(status_path.read_text()) if status_path.exists() else {}
    identity = [str(args.root), str(args.record), args.main, args.binary]
    if state and state["identity"] != identity:
        raise RuntimeError("workflow identity changed")
    state["identity"] = identity
    if not args.record.exists():
        arguments = [args.binary, "record", "-out", str(args.record), "-main", args.main,
                     "-jobs", "4", "-isolate", package]
        metadata = timing_metadata(args.root, environment, arguments)
        metadata["build_flags"] = "go test -c; -test.v=test2json -test.timeout=60m; four workers, markdownblocks last and alone"
        state["phase"] = "reference"
        state.setdefault("reference_sessions", []).append(metadata)
        atomic_json(status_path, state)
        with (args.output / "reference.log").open("a") as log:
            result = subprocess.run(arguments, cwd=args.root, env=environment, stdout=log, stderr=log)
        metadata.update(ended=time.time(), load_after=Path("/proc/loadavg").read_text().strip(), exit=result.returncode)
        atomic_json(status_path, state)
        if result.returncode:
            raise RuntimeError("reference failed; all attempts retained; explicit diagnosis required")
    record = json.loads(args.record.read_text())
    if not record.get("Complete") or record["Commit"] != args.main:
        raise RuntimeError("reference is not complete at fixed main")
    if "calibration" not in state:
        state["phase"] = "markdownblocks_without_observer"
        atomic_json(status_path, state)
        destination = args.output / "markdown-plain"
        destination.mkdir(exist_ok=True)
        plain_environment = dict(environment, ADAMIC_GATE_UNCACHED="1")
        metadata = timing_metadata(args.root, plain_environment,
                                   ["go test -c", package, "direct binary -test.v=test2json -test.timeout=60m"])
        metadata["build_flags"] = "go test -c; -test.v=test2json -test.timeout=60m; one package alone, no observer"
        started = time.monotonic()
        result = run_package(package, args.root / "stage1/cohere/markdownblocks", args.root,
                             destination, plain_environment)
        metadata.update(wall_seconds=time.monotonic()-started, ended=time.time(),
                        load_after=Path("/proc/loadavg").read_text().strip())
        observed_events = record["Packages"][package]["Events"]
        observed = json.loads(Path(observed_events.removesuffix(".jsonl")+".timing.json").read_text())
        # test2json's package elapsed excludes compilation in both runs.
        terminal = [json.loads(line) for line in Path(result["events"]).read_text().splitlines()
                    if not json.loads(line).get("Test") and json.loads(line)["Action"] in ("pass", "fail")]
        plain_seconds = terminal[-1]["Elapsed"]
        state["calibration"] = {"plain": result, "metadata": metadata, "observed": observed,
                                "plain_process_seconds": plain_seconds,
                                "observer_overhead_seconds": observed["wall_seconds"]-plain_seconds,
                                "identical": canonical_events(result["events"]) == canonical_events(observed_events)}
        atomic_json(status_path, state)
    # Markdownblocks is always selected. Its benchmark throughput and generated
    # temporary paths vary, so calibration compares test names and verdicts;
    # strict output comparison remains mandatory for every skipped package.
    calibration = state["calibration"]
    plain_events = canonical_events(calibration["plain"]["events"])
    observed_events = canonical_events(record["Packages"][package]["Events"])
    calibration["verdicts_identical"] = calibration_verdicts(plain_events) == calibration_verdicts(observed_events)
    if not calibration["identical"]:
        atomic_json(args.output / "calibration-differences.json",
                    {name: {"observed": observed_events.get(name), "plain": plain_events.get(name)}
                     for name in sorted(set(plain_events) | set(observed_events))
                     if plain_events.get(name) != observed_events.get(name)})
    atomic_json(status_path, state)
    if calibration["plain"]["exit"] or not calibration["verdicts_identical"]:
        raise RuntimeError("markdownblocks calibration failed or changed test names/verdicts")
    state["phase"] = "five_shapes"
    atomic_json(status_path, state)
    arguments = [sys.executable, "-B", str(Path(__file__).with_name("proof.py")),
                 "--root", str(args.root), "--record", str(args.record),
                 "--output", str(args.output / "branches"), "--binary", args.binary]
    with (args.output / "branches.log").open("a") as log, (args.output / "branches.stderr").open("a") as errors:
        result = subprocess.run(arguments, cwd=args.root, env=environment, stdout=log, stderr=errors)
    state.update(phase="finished", exit=result.returncode)
    atomic_json(status_path, state)
    if result.returncode:
        raise RuntimeError("five-shape proof failed; see durable state and logs")


if __name__ == "__main__":
    main()
