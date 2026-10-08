#!/usr/bin/env python3
"""CPU-weighted full-gate planning, SSH coordination and fail-closed merging."""
import argparse
from collections import Counter
from concurrent.futures import ThreadPoolExecutor
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import time


def plan(packages, weights, boxes):
    if not boxes or any(cpus <= 0 for cpus in boxes.values()):
        raise ValueError("boxes need positive CPU counts")
    if len(packages) != len(set(packages)) or any(not re.fullmatch(r"[A-Za-z0-9_.~+/-]+", name) for name in packages):
        raise ValueError("duplicate or invalid package names")
    if any(not 0 <= value < float("inf") for value in weights.values()):
        raise ValueError("weights must be finite and nonnegative")
    shards = [{"box": box, "cpus": cpus, "packages": [], "estimated_seconds": 0.0}
              for box, cpus in sorted(boxes.items())]
    for package in sorted(packages, key=lambda name: (-weights.get(name, 1.0), name)):
        seconds = weights.get(package, 1.0)
        shard = min(shards, key=lambda row: ((row["estimated_seconds"] + seconds) / row["cpus"], row["box"]))
        shard["packages"].append(package)
        shard["estimated_seconds"] += seconds
    owner = min(shards, key=lambda row: (-row["cpus"], row["box"]))["box"]
    for index, shard in enumerate(shards):
        shard.update(index=index, once_steps=shard["box"] == owner,
                     estimated_wall_seconds=shard["estimated_seconds"] / shard["cpus"])
    return {"packages": sorted(packages), "weights": dict(sorted(weights.items())), "once_box": owner, "shards": shards}


def event_counts(paths, strict=True):
    counts = Counter()
    for path in paths:
        with open(path) as handle:
            for line in handle:
                try:
                    event = json.loads(line)
                except ValueError:
                    if strict:
                        raise
                    continue
                if event.get("Test") is not None and event.get("Action") in ("pass", "fail", "skip"):
                    counts[(event["Package"], event["Action"])] += 1
    return counts


def write_json(path, value):
    temporary = Path(str(path) + ".tmp")
    temporary.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n")
    temporary.replace(path)


def merge(layout, root, sha, tools, final=True, census=None, wall=0):
    root = Path(root)
    root.mkdir(parents=True, exist_ok=True)
    assignments = [package for shard in layout["shards"] for package in shard["packages"]]
    if sorted(assignments) != sorted(layout["packages"]) or len(assignments) != len(set(assignments)):
        raise ValueError("plan does not cover every package exactly once")
    failures, reports = [], []
    totals = Counter()
    steps, exits, planned = {}, {}, []
    logs = []
    for shard in layout["shards"]:
        label = "shard-%d@%s" % (shard["index"], shard["box"])
        directory = root / "shards" / str(shard["index"])
        report_path = directory / "full.json"
        report = None
        try:
            if report_path.exists():
                report = json.loads(report_path.read_text())
                if report["sha"] != sha or report["tools_sha"] != tools or sorted(report["package_list"]) != sorted(shard["packages"]):
                    raise ValueError("wrong source, tools or package assignment")
                expected = {"coverage", "build", "vet", "tests", "wasi", "stage3", "catalog"} if shard["once_steps"] else {"tests"}
                if set(report["planned_stages"]) != expected:
                    raise ValueError("incorrect planned stages")
                if not report.get("finished") or any(report.get("stages_exit", {}).get(stage) != 0 for stage in report["planned_stages"]):
                    if not report.get("failure"):
                        raise ValueError("unfinished or unreported stage")
                if report.get("failure"):
                    failures.append(dict(report["failure"], step=label + "/" + report["failure"]["step"]))
                reports.append(dict(report, shard=shard["index"], box=shard["box"]))
                for stage, seconds in report["steps_seconds"].items():
                    steps[label + "/" + stage] = seconds
                for stage in report["planned_stages"]:
                    planned.append(label + "/" + stage)
                    exits[label + "/" + stage] = report["stages_exit"].get(stage)
            elif (directory / "status.txt").exists() and (directory / "status.txt").read_text().startswith("red:"):
                detail = (directory / "first-failure.txt").read_text() if (directory / "first-failure.txt").exists() else (directory / "status.txt").read_text()
                failures.append({"step": label, "detail": detail, "after_seconds": wall})
            if (directory / "transport-error.txt").exists():
                raise ValueError((directory / "transport-error.txt").read_text())
            if final and report is None:
                raise ValueError("missing shard report (box down, timeout or runner did not finish)")
            log = directory / "test.jsonl"
            if log.exists():
                counts = event_counts([log], strict=final)
                tally = Counter()
                for (_, action), count in counts.items():
                    tally[action] += count
                if final and tally["fail"] and not (report or {}).get("failure"):
                    failures.append({"step": label + "/tests", "detail": "failing test events in shard log", "after_seconds": wall})
                totals.update(tally)
                logs.append(log)
                if final and report and any(report[action] != tally[action] for action in ("pass", "fail", "skip")):
                    raise ValueError("reported counts disagree with test.jsonl")
            elif report:
                raise ValueError("missing test.jsonl")
        except (ValueError, KeyError, OSError, TypeError) as error:
            if not final and not (directory / "transport-error.txt").exists():
                continue
            failures.append({"step": label, "detail": str(error), "after_seconds": wall})
    with open(root / "test.jsonl.tmp", "w") as output:
        for log in logs:
            with open(log) as handle:
                for line in handle:
                    output.write(line)
    (root / "test.jsonl.tmp").replace(root / "test.jsonl")
    if final:
        planned.append("census")
        steps["census"] = (census or {}).get("steps_seconds", {}).get("census", 0)
        if census and (census.get("sha") != sha or census.get("tools_sha") != tools or not census.get("finished")):
            failures.append({"step": "shard-census@" + layout["once_box"], "detail": "wrong census identity or unfinished report", "after_seconds": wall})
        exits["census"] = (census or {}).get("stages_exit", {}).get("census")
        if exits["census"] != 0 or (census or {}).get("failure"):
            failures.append({"step": "shard-census@" + layout["once_box"], "detail": (census or {}).get("failure") or "missing census report", "after_seconds": wall})
    # Preserve the first observed red while later shards continue for triage.
    first_path = root / "failure.json"
    if first_path.exists():
        failure = json.loads(first_path.read_text())
    else:
        failure = failures[0] if failures else None
        if failure:
            write_json(first_path, failure)
    result = {"sha": sha, "base": sha, "tools_sha": tools, "packages": "all", "package_list": layout["packages"],
              "uncached_tests": True, "finished": final, "failure": failure,
              "pass": totals["pass"], "fail": totals["fail"], "skip": totals["skip"],
              "wall_seconds": round(wall, 1), "steps_seconds": steps, "planned_stages": planned, "stages_exit": exits,
              "build_ok": any(row.get("build_ok") for row in reports), "vet_ok": any(row.get("vet_ok") for row in reports),
              "required_input_skips": (census or {}).get("required_input_skips", []),
              "unclassified_skips": (census or {}).get("unclassified_skips", []), "shards": reports, "plan": layout}
    summary = " ".join("%s=%.1fs" % item for item in steps.items())
    if failure:
        status = "red: %s full gate, first failure at %s after %.1f s (%s), %d fail, %d pass" % (sha, failure["step"], failure["after_seconds"], summary, totals["fail"], totals["pass"])
        (root / "first-failure.txt").write_text(str(failure["detail"]) + "\n")
    elif final:
        status = "green: %s full gate in %.1f s (%s), %d packages, %d pass, %d skip, smoke 0 fixtures" % (sha, wall, summary, len(layout["packages"]), totals["pass"], totals["skip"])
    else:
        status = "running: full gate of %s, %d/%d shards reported" % (sha, len(reports), len(layout["shards"]))
    (root / "status.txt.tmp").write_text(status + "\n")
    (root / "status.txt.tmp").replace(root / "status.txt")
    if final:
        write_json(root / "full.json", result)
    return result


SETUP = '''set -euo pipefail
source ~/adamic-tools/env.sh
mkdir -p -m 1777 "${TMPDIR:-/tmp}"
exec 9> ~/full-gate/lock
flock 9
git -C ~/full-gate/tools fetch -q origin "$2"
git -C ~/full-gate/tools switch -q --detach "$2"
git -C ~/full-gate/tree fetch -q origin "$1"
git -C ~/full-gate/tree switch -q --detach "$1"
git -C ~/full-gate/tree submodule update -q --init --recursive
'''
SSH = ["ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=15", "-o", "ServerAliveInterval=15", "-o", "ServerAliveCountMax=3"]


def remote(box, script, arguments, timeout=300):
    return subprocess.run(SSH + [box, "bash -s -- " + " ".join(shlex.quote(str(arg)) for arg in arguments)],
                          input=script, text=True, capture_output=True, timeout=timeout, check=True).stdout


def coordinate(args):
    if not re.fullmatch(r"[0-9a-f]{40}", args.sha) or not re.fullmatch(r"[0-9a-f]{40}", args.tools):
        raise ValueError("source and tools must be full commit SHAs")
    root = Path(args.out).resolve()
    if not re.fullmatch(r"[A-Za-z0-9_.-]+", root.name):
        raise ValueError("output directory basename must be a simple run identifier")
    root.mkdir(parents=True, exist_ok=True)
    boxes = args.boxes.split()
    if len(boxes) != len(set(boxes)) or any(not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]*", box) for box in boxes):
        raise ValueError("box aliases must be distinct simple SSH names")
    started = time.monotonic()
    with ThreadPoolExecutor(max_workers=len(boxes)) as pool:
        probes = {box: pool.submit(remote, box, SETUP + '''cd ~/full-gate/tree
nproc --all
go list ./...
printf '\\nWEIGHTS\\n'
cat ~/full-gate/weights.txt
''', [args.sha, args.tools]) for box in boxes}
        discovered, weights_text, cpus = None, None, {}
        for box, future in probes.items():
            try:
                listing, weights = future.result().split("\nWEIGHTS\n", 1)
                cpu, *packages = listing.splitlines()
                if discovered is not None and (set(discovered) != set(packages) or weights_text != weights):
                    raise ValueError("package list or weights differ across boxes")
                discovered, weights_text = packages, weights
                cpus[box] = int(cpu)
            except Exception as error:
                failure = {"step": "setup@" + box, "detail": str(error), "after_seconds": round(time.monotonic() - started, 1)}
                (root / "status.txt").write_text("red: %s full gate, first failure at %s\n" % (args.sha, failure["step"]))
                (root / "first-failure.txt").write_text(str(error) + "\n")
                write_json(root / "full.json", {"sha": args.sha, "finished": True, "failure": failure})
                return
        weights = {name: float(seconds) for seconds, name in (line.split() for line in weights_text.splitlines() if line.strip())}
        layout = plan(discovered, weights, cpus)
        write_json(root / "plan.json", layout)
        print(json.dumps(layout, indent=2), flush=True)
        remote_out = "full-gate/out/" + root.name

        def execute(shard):
            directory = root / "shards" / str(shard["index"])
            directory.mkdir(parents=True, exist_ok=True)
            packages = "\n".join(shard["packages"]) + "\n"
            script = SETUP + 'mkdir -p ~/' + remote_out + '\ncat > ~/' + remote_out + '/packages.txt <<\'PACKAGES\'\n' + packages + 'PACKAGES\n'
            script += 'timeout ' + str(args.timeout) + 's python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --tree ~/full-gate/tree --sha "$1" --base "$1" --tools ~/full-gate/tools --weights ~/full-gate/weights.txt --out ~/' + remote_out + ' --packages-file ~/' + remote_out + '/packages.txt' + (' --once-steps' if shard["once_steps"] else '') + ' > ~/' + remote_out + '/driver.log 2>&1\n'
            try:
                remote(shard["box"], script, [args.sha, args.tools], timeout=args.timeout + 600)
            except subprocess.CalledProcessError as error:
                # A gate's red exits 1; collect its report before deciding transport failed.
                if error.returncode != 1:
                    (directory / "transport-error.txt").write_text(str(error) + "\n" + error.stderr)
            except Exception as error:
                (directory / "transport-error.txt").write_text(str(error))

        futures = [pool.submit(execute, shard) for shard in layout["shards"]]
        while True:
            done = all(future.done() for future in futures)
            for shard in layout["shards"]:
                directory = root / "shards" / str(shard["index"])
                directory.mkdir(parents=True, exist_ok=True)
                # Copy into a staging directory; a failed copy cannot overwrite prior evidence.
                staging = directory / "snapshot"
                staging.mkdir(exist_ok=True)
                try:
                    copy = subprocess.run(["scp", "-q", "-r", "-o", "BatchMode=yes", "-o", "ConnectTimeout=15", "-o", "ServerAliveInterval=15", "-o", "ServerAliveCountMax=3",
                                       shard["box"] + ":" + remote_out + "/.", str(staging)], capture_output=True, timeout=120)
                except subprocess.TimeoutExpired:
                    (directory / "transport-error.txt").write_text("timed out collecting shard")
                    continue
                if copy.returncode == 0:
                    for path in staging.iterdir():
                        if path.is_file():
                            path.replace(directory / path.name)
                elif done:
                    (directory / "transport-error.txt").write_text("cannot collect shard: " + copy.stderr.decode(errors="replace"))
            merge(layout, root, args.sha, args.tools, final=False, wall=time.monotonic() - started)
            if done:
                for future in futures:
                    future.result()
                break
            time.sleep(15)
        census = None
        try:
            owner = layout["once_box"]
            census_out = remote_out + "-census"
            remote(owner, 'mkdir -p ~/' + census_out + '\n', [])
            subprocess.run(["scp", "-q", str(root / "test.jsonl"), owner + ":" + census_out + "/test.jsonl"], check=True, timeout=120)
            try:
                remote(owner, SETUP + 'python3 ~/full-gate/tools/cloud/fast-gate/run.py --full --census-only --tree ~/full-gate/tree --sha "$1" --base "$1" --tools ~/full-gate/tools --out ~/' + census_out + '\n', [args.sha, args.tools], timeout=1800)
            except subprocess.CalledProcessError:
                pass
            destination = root / "census"
            subprocess.run(["scp", "-q", "-r", owner + ":" + census_out, str(destination)], check=True, timeout=120)
            census = json.loads((destination / "full.json").read_text())
        except Exception as error:
            (root / "census-error.txt").write_text(str(error))
        merge(layout, root, args.sha, args.tools, census=census, wall=time.monotonic() - started)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    compare = commands.add_parser("compare", help="compare raw pass/fail/skip test counts per package")
    compare.add_argument("--whole", nargs="+", required=True)
    compare.add_argument("--sharded", nargs="+", required=True)
    launch = commands.add_parser("run")
    launch.add_argument("--boxes", required=True)
    launch.add_argument("--sha", required=True)
    launch.add_argument("--tools", required=True)
    launch.add_argument("--out", required=True)
    launch.add_argument("--timeout", type=int, default=int(os.environ.get("ADAMIC_FULL_GATE_TIMEOUT", "86400")))
    args = parser.parse_args()
    if args.command == "run" and args.timeout <= 0:
        parser.error("timeout must be positive")
    if args.command == "compare":
        whole, sharded = event_counts(args.whole), event_counts(args.sharded)
        differences = [(key, whole[key], sharded[key]) for key in sorted(whole.keys() | sharded.keys()) if whole[key] != sharded[key]]
        for key, left, right in differences:
            print("%s %s: whole=%d sharded=%d" % (*key, left, right))
        print("different" if differences else "equal: raw pass/fail/skip test events per package")
        return bool(differences)
    try:
        coordinate(args)
    except Exception as error:
        root = Path(args.out)
        root.mkdir(parents=True, exist_ok=True)
        (root / "status.txt").write_text("red: %s full gate, first failure at coordinator: %s\n" % (args.sha, error))
        (root / "first-failure.txt").write_text(str(error) + "\n")
        write_json(root / "full.json", {"sha": args.sha, "finished": True, "failure": {"step": "coordinator", "detail": str(error)}})
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
