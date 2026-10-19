#!/usr/bin/env python3
"""Clock emit's two heads on #nda35wb: clang seconds and peak memory for four stage1 programs.

Usage:

    python3 -I docs/tsc-generated-c/emit-split-clock/clock.py --clang CLANG --out DIRECTORY \
        --tree LABEL=WORKTREE [--tree LABEL=WORKTREE ...]

Each WORKTREE is a detached checkout holding its own built `adamic` binary (`go build -o WORKTREE/adamic
./cmd/adamic`). For each program the script emits C with that tree's `adamic c main.ts` (run from the
program's directory, as measure.py does), then times clang on the emitted file one compile at a time with
measure.py's FLAGS, `-I WORKTREE/internal/native/runtime -c`, at -O0, -O1 and -O2, three repeats each,
and keeps the median. Repeats are interleaved across trees (tree a, b, c, d, then again) so a slow drift
of the box lands on every tree alike. At -O2 on tsprinter it also runs once per tree with
-ftime-trace -ftime-trace-granularity=50.

FLAGS, functions(), clang_once() and time_trace_top() are copied from docs/tsc-generated-c/measure.py on
origin/compiler/tsc-generated-c-table-2; clang_once takes the runtime directory as an argument here,
since each tree carries its own runtime.
"""
import argparse
import json
import os
import re
import shutil
import statistics
import subprocess
import sys
import tempfile
import time

PROGRAMS = ["stage1/cohere/tsprinter/main.ts", "stage1/cohere/json/main.ts", "stage1/cohere/yaml/main.ts",
            "stage1/cohere/graphql/printer/main.ts"]
LEVELS = ["-O0", "-O1", "-O2"]
REPEATS = 3

# Copied from measure.py.
FLAGS = ["-std=c11", "-Wall", "-Wextra", "-Werror", "-Wcast-function-type-strict", "-pedantic",
         "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function",
         "-Wno-unused-parameter", "-Wno-self-assign", "-ffp-contract=off", "-fno-optimize-sibling-calls",
         "-pthread"]

FUNCTION = re.compile(r"^(?!#)(?!typedef)[A-Za-z_][\w \t\*]*?\b(\w+)\s*\([^;]*\)\s*\{\s*$")


def loadavg():
    with open("/proc/loadavg") as handle:
        return handle.read().split()[:3]


def functions(c_text):
    """Copied from measure.py: top-level C function definitions as (name, lines)."""
    found, name, start = [], None, 0
    lines = c_text.split("\n")
    for index, line in enumerate(lines):
        if name is None:
            match = FUNCTION.match(line)
            if match and not line.startswith(("struct ", "union ", "enum ")):
                name, start = match.group(1), index
        elif line == "}" or line.startswith("};"):
            found.append((name, index - start + 1))
            name = None
    return found


def clang_once(clang, runtime, c_path, level, directory, trace=False):
    """Copied from measure.py, with the runtime directory passed in."""
    arguments = [clang] + FLAGS + [level, "-I", runtime, "-c", c_path, "-o", os.path.join(directory, "unit.o")]
    if trace:
        arguments += ["-ftime-trace", "-ftime-trace-granularity=50"]
    before = loadavg()
    process = subprocess.run(["/usr/bin/time", "-v"] + arguments, cwd=directory, capture_output=True, text=True)
    after = loadavg()
    report = process.stderr
    def field(label):
        match = re.search(r"^\s*" + re.escape(label) + r": (.+)$", report, re.M)
        return match.group(1).strip() if match else ""
    wall = field("Elapsed (wall clock) time (h:mm:ss or m:ss)")
    parts = [float(part) for part in wall.split(":")] if wall else [0.0]
    seconds = 0.0
    for part in parts:
        seconds = seconds * 60 + part
    row = {
        "level": level,
        "exit": process.returncode,
        "wall_seconds": round(seconds, 3),
        "user_seconds": float(field("User time (seconds)") or 0),
        "system_seconds": float(field("System time (seconds)") or 0),
        "max_rss_kb": int(field("Maximum resident set size (kbytes)") or 0),
        "object_bytes": os.path.getsize(os.path.join(directory, "unit.o")) if process.returncode == 0 else 0,
        "load_before": before,
        "load_after": after,
    }
    if process.returncode != 0:
        row["message"] = report.split("\tCommand being timed")[0].strip()[:2000]
    return row


def time_trace_top(directory):
    """Copied from measure.py."""
    path = os.path.join(directory, "unit.json")
    with open(path) as handle:
        events = json.load(handle)["traceEvents"]
    totals = [event for event in events if event.get("name", "").startswith("Total ") and "dur" in event]
    totals.sort(key=lambda event: event["dur"], reverse=True)
    passes, by_function = {}, {}
    for event in events:
        detail = event.get("args", {}).get("detail")
        if event.get("ph") != "X" or detail is None:
            continue
        if event.get("name") in ("RunPass", "RunLoopPass"):
            passes[detail] = passes.get(detail, 0) + event["dur"] / 1e6
        elif event.get("name") == "OptFunction":
            by_function[detail] = by_function.get(detail, 0) + event["dur"] / 1e6
    top = lambda table, count: [{"name": name, "seconds": round(seconds, 3)} for name, seconds in sorted(table.items(), key=lambda item: -item[1])[:count]]
    return {"outline": [{"phase": event["name"][len("Total "):], "seconds": round(event["dur"] / 1e6, 3)} for event in totals[:12]],
            "passes": top(passes, 8), "functions": top(by_function, 3)}


def box_facts(clang):
    with open("/proc/cpuinfo") as handle:
        cpu = next(line.split(":", 1)[1].strip() for line in handle if line.startswith("model name"))
    with open("/proc/meminfo") as handle:
        memory_kb = int(handle.readline().split()[1])
    lscpu = subprocess.run(["lscpu"], capture_output=True, text=True).stdout
    model = next((line.split(":", 1)[1].strip() for line in lscpu.split("\n") if line.startswith("Model name")), "")
    return {"nproc": int(subprocess.run(["nproc"], capture_output=True, text=True).stdout.strip()),
            "lscpu_model_name": model, "cpuinfo_model_name": cpu, "memory_kb": memory_kb,
            "kernel": os.uname().release, "clang": clang,
            "clang_version": subprocess.run([clang, "--version"], capture_output=True, text=True).stdout.strip(),
            "go_version": subprocess.run(["go", "version"], capture_output=True, text=True).stdout.strip(),
            "flags": FLAGS}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--clang", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--tree", action="append", required=True, help="LABEL=WORKTREE")
    arguments = parser.parse_args()
    os.makedirs(os.path.join(arguments.out, "c"), exist_ok=True)
    trees = []
    for item in arguments.tree:
        label, path = item.split("=", 1)
        commit = subprocess.run(["git", "-C", path, "rev-parse", "HEAD"], capture_output=True, text=True).stdout.strip()
        trees.append({"label": label, "path": os.path.abspath(path), "commit": commit})

    box = box_facts(arguments.clang)
    box["load_at_start"] = loadavg()
    rows, traces = [], []
    for program in PROGRAMS:
        program_load_before = loadavg()
        print(f"{program} load {program_load_before}", flush=True)
        units = {}
        for tree in trees:
            entry = os.path.join(tree["path"], program)
            started = time.monotonic()
            process = subprocess.run([os.path.join(tree["path"], "adamic"), "c", entry], cwd=os.path.dirname(entry),
                                     capture_output=True)
            row = {"tree": tree["label"], "commit": tree["commit"], "program": program,
                   "adamic_exit": process.returncode, "adamic_seconds": round(time.monotonic() - started, 3),
                   "load_before_program": program_load_before}
            rows.append(row)
            if process.returncode != 0:
                row["adamic_stderr"] = process.stderr.decode("utf-8", "replace")[:2000]
                print(f"  {tree['label']}: adamic c failed: {row['adamic_stderr'][:200]}", flush=True)
                continue
            text = process.stdout.decode("utf-8", "replace")
            found = functions(text)
            largest = max(found, key=lambda item: item[1]) if found else ("", 0)
            row.update({"c_bytes": len(process.stdout), "c_lines": text.count("\n"), "c_functions": len(found),
                        "largest_function": largest[0], "largest_function_lines": largest[1],
                        "main_functions": [{"name": name, "lines": lines} for name, lines in found
                                           if name == "main" or name.startswith("adamic_main_")],
                        "clang": {}})
            name = tree["label"] + "_" + program.replace("/", "_").replace("stage1_cohere_", "")
            with open(os.path.join(arguments.out, "c", name + ".c"), "wb") as handle:
                handle.write(process.stdout)
            directory = tempfile.mkdtemp(prefix="clock-")
            unit = os.path.join(directory, "unit.c")
            with open(unit, "wb") as handle:
                handle.write(process.stdout)
            units[tree["label"]] = (row, tree, directory, unit)

        for level in LEVELS:
            for repeat in range(REPEATS):
                for label, (row, tree, directory, unit) in units.items():
                    runs = row["clang"].setdefault(level, {"runs": []})["runs"]
                    if runs and runs[-1]["exit"] != 0:
                        continue
                    sample = clang_once(arguments.clang, os.path.join(tree["path"], "internal", "native", "runtime"),
                                        unit, level, directory)
                    sample["repeat"] = repeat + 1
                    runs.append(sample)
                    print(f"  {label} {level} #{repeat + 1}: {sample['wall_seconds']}s {sample['max_rss_kb']} KB "
                          f"exit {sample['exit']} load {sample['load_before']}", flush=True)
            for label, (row, tree, directory, unit) in units.items():
                summary = row["clang"][level]
                good = [run for run in summary["runs"] if run["exit"] == 0]
                summary["all_exit_zero"] = len(good) == len(summary["runs"]) == REPEATS
                if good:
                    summary["median_wall_seconds"] = statistics.median(run["wall_seconds"] for run in good)
                    summary["median_max_rss_mb"] = round(statistics.median(run["max_rss_kb"] for run in good) / 1024, 1)
                    summary["raw_wall_seconds"] = [run["wall_seconds"] for run in summary["runs"]]
                    summary["raw_max_rss_mb"] = [round(run["max_rss_kb"] / 1024, 1) for run in summary["runs"]]

        if program == "stage1/cohere/tsprinter/main.ts":
            for label, (row, tree, directory, unit) in units.items():
                sample = clang_once(arguments.clang, os.path.join(tree["path"], "internal", "native", "runtime"),
                                    unit, "-O2", directory, trace=True)
                trace = {"tree": label, "program": program, "run": sample}
                if sample["exit"] == 0:
                    trace["phases"] = time_trace_top(directory)
                traces.append(trace)
                top = trace.get("phases", {}).get("functions", [{}])[0]
                print(f"  {label} -O2 trace: {sample['wall_seconds']}s slowest function {top}", flush=True)

        for label, (row, tree, directory, unit) in units.items():
            shutil.rmtree(directory)
        program_load_after = loadavg()
        for row in rows:
            if row["program"] == program:
                row["load_after_program"] = program_load_after

    box["load_at_end"] = loadavg()
    with open(os.path.join(arguments.out, "rows.json"), "w") as handle:
        json.dump({"trees": trees, "box": box, "rows": rows, "time_trace": traces}, handle, indent=1)
        handle.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
