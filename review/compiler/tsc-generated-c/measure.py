#!/usr/bin/env python3
"""Measure the C Adamic emits per file: bytes, lines, units, largest function, clang seconds and memory.

Usage (from the repository root, after building cmd/adamic and running stage3/apply.sh):

    python3 -I review/compiler/tsc-generated-c/measure.py \
        --adamic BIN --clang CLANG --adapted ADAPTED_TREE --meter METER_JSON --out DIRECTORY [--only N]

It compiles every file under ADAPTED_TREE/src/compiler with `adamic c` (each file on its own, as the
meter's lowering does), then every stage1 main.ts as labeled proxies, and times clang on each emitted
C file one at a time at -O0, -O1 and -O2 under /usr/bin/time -v. Results go to OUT/rows.json and
OUT/files.csv; emitted C stays in OUT/c (not committed).
"""
import argparse
import csv
import json
import math
import os
import re
import subprocess
import sys
import tempfile
import time

REPOSITORY = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
RUNTIME = os.path.join(REPOSITORY, "internal", "native", "runtime")

# native.Flags for an ordinary release build on Linux (internal/native/native.go), without -O, which
# the measurement varies. Split builds and ThinLTO are opt-in and not measured here.
FLAGS = ["-std=c11", "-Wall", "-Wextra", "-Werror", "-Wcast-function-type-strict", "-pedantic",
         "-Wno-unused-variable", "-Wno-unused-but-set-variable", "-Wno-unused-function",
         "-Wno-unused-parameter", "-Wno-self-assign", "-ffp-contract=off", "-fno-optimize-sibling-calls",
         "-pthread"]

FUNCTION = re.compile(r"^(?!#)(?!typedef)[A-Za-z_][\w \t\*]*?\b(\w+)\s*\([^;]*\)\s*\{\s*$")
IMPORT = re.compile(r"""(?:\bfrom\s+|\bimport\s+|\bimport\s*\(\s*)["'](\.{1,2}/[^"']+)["']""")


def loadavg():
    with open("/proc/loadavg") as handle:
        return handle.read().split()[:3]


def closure(entry):
    """The relative-import closure of entry: what `adamic c entry` reads besides the library."""
    seen, stack = set(), [os.path.abspath(entry)]
    while stack:
        path = stack.pop()
        if path in seen:
            continue
        seen.add(path)
        with open(path, encoding="utf-8") as handle:
            text = handle.read()
        for specifier in IMPORT.findall(text):
            base = os.path.normpath(os.path.join(os.path.dirname(path), specifier))
            for candidate in (base, base + ".ts", base + ".a", base[:-3] + ".ts" if base.endswith(".js") else "",
                              os.path.join(base, "index.ts"), os.path.join(base, "index.a")):
                if candidate and os.path.isfile(candidate) and not candidate.endswith(".d.ts"):
                    stack.append(candidate)
                    break
    return sorted(seen)


def source_size(paths):
    total_bytes = total_lines = 0
    for path in paths:
        with open(path, "rb") as handle:
            data = handle.read()
        total_bytes += len(data)
        total_lines += data.count(b"\n")
    return total_bytes, total_lines


def functions(c_text):
    """Top-level C function definitions: (name, lines). A definition opens at column 0 and closes at a
    column-0 brace, which is how emit.go writes every function."""
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


def clang_once(clang, c_path, level, directory, trace=False):
    arguments = [clang] + FLAGS + [level, "-I", RUNTIME, "-c", c_path, "-o", os.path.join(directory, "unit.o")]
    if trace:
        arguments += ["-ftime-trace", "-ftime-trace-granularity=500"]
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
    path = os.path.join(directory, "unit.json")
    with open(path) as handle:
        events = json.load(handle)["traceEvents"]
    totals = [event for event in events if event.get("name", "").startswith("Total ") and "dur" in event]
    totals.sort(key=lambda event: event["dur"], reverse=True)
    # "Total ExecuteCompiler" and "Total Frontend"/"Total Backend" nest the rest; keep them, they are the
    # outline, and list the leaf phases after them.
    return [{"phase": event["name"][len("Total "):], "seconds": round(event["dur"] / 1e6, 3)} for event in totals[:12]]


def measure_program(arguments, label, kind, entry, sources, meter_reason, out):
    name = label.replace("/", "_")
    c_path = os.path.join(out, "c", name + ".c")
    started = time.monotonic()
    process = subprocess.run([arguments.adamic, "c", entry], cwd=os.path.dirname(entry), capture_output=True)
    lowering_seconds = round(time.monotonic() - started, 3)
    source_bytes, source_lines = source_size(sources)
    own_bytes, own_lines = source_size([entry])
    row = {"file": label, "kind": kind, "own_ts_bytes": own_bytes, "own_ts_lines": own_lines,
           "closure_files": len(sources), "closure_ts_bytes": source_bytes, "closure_ts_lines": source_lines,
           "adamic_exit": process.returncode, "adamic_seconds": lowering_seconds}
    if meter_reason is not None:
        row["meter_reason"] = meter_reason
    if process.returncode != 0:
        lines = process.stderr.decode("utf-8", "replace").strip().split("\n")
        row["first_reason"] = lines[0][:500] if lines else "(no message)"
        return row
    with open(c_path, "wb") as handle:
        handle.write(process.stdout)
    text = process.stdout.decode("utf-8", "replace")
    found = functions(text)
    largest = max(found, key=lambda item: item[1]) if found else ("", 0)
    nonmain = [item for item in found if item[0] != "main"]
    row.update({
        "c_bytes": len(process.stdout),
        "c_lines": text.count("\n"),
        "c_functions": len(found),
        # The default native build compiles main.c as one translation unit (native.go buildOne).
        "units": 1,
        # What the opt-in split build (units.go) would make: state.c, groups of sixteen functions, main.c.
        "split_units_if_opted_in": 2 + math.ceil(len(nonmain) / 16),
        "largest_function": largest[0],
        "largest_function_lines": largest[1],
        "clang": [],
    })
    with tempfile.TemporaryDirectory(prefix="tsc-c-") as directory:
        unit = os.path.join(directory, "unit.c")
        with open(unit, "wb") as handle:
            handle.write(process.stdout)
        for level in ("-O0", "-O1", "-O2"):
            samples = [clang_once(arguments.clang, unit, level, directory)]
            # Small units repeat to beat timer and startup noise; the minimum is reported.
            repeats = arguments.repeats if row["c_bytes"] < 300_000 else 1
            while len(samples) < repeats and samples[-1]["exit"] == 0:
                samples.append(clang_once(arguments.clang, unit, level, directory))
            best = min(samples, key=lambda sample: sample["wall_seconds"])
            best["samples_wall_seconds"] = [sample["wall_seconds"] for sample in samples]
            row["clang"].append(best)
            print(f"  {label} {level}: {best['wall_seconds']}s {best['max_rss_kb']} KB exit {best['exit']}", flush=True)
            if best["exit"] != 0:
                break
    return row


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--adamic", required=True)
    parser.add_argument("--clang", required=True)
    parser.add_argument("--adapted", required=True)
    parser.add_argument("--meter", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--repeats", type=int, default=3)
    parser.add_argument("--only", type=int, default=0, help="measure only the first N proxies, smallest C first")
    parser.add_argument("--skip-proxies", action="store_true")
    arguments = parser.parse_args()
    os.makedirs(os.path.join(arguments.out, "c"), exist_ok=True)

    compiler = os.path.join(arguments.adapted, "src", "compiler")
    with open(arguments.meter) as handle:
        meter = json.load(handle)
    # The meter ranks reasons by count; a file's first reason is the highest-ranked one naming it.
    meter_first = {}
    for reason in meter["reasons"]:
        for file in reason["files"]:
            meter_first.setdefault(file, f"{reason['kind']}: {reason['reason']} (e.g. {reason['example']})")

    rows = []
    tsc_files = sorted(os.path.relpath(os.path.join(directory, file), compiler)
                       for directory, _, files in os.walk(compiler) for file in files if file.endswith(".ts"))
    for file in tsc_files:
        entry = os.path.join(compiler, file)
        print(f"tsc {file}", flush=True)
        rows.append(measure_program(arguments, file, "tsc", entry, closure(entry),
                                    meter_first.get(file, "none of its own (reaches lowering only if its import closure is clean)"),
                                    arguments.out))
    if not arguments.skip_proxies:
        entries = []
        for directory, _, files in os.walk(os.path.join(REPOSITORY, "stage1")):
            if "/gaps" in directory or "main.ts" not in files:
                continue
            entries.append(os.path.join(directory, "main.ts"))
        entries.sort()
        # Order proxies by their closure size so a partial run measures the small end first.
        entries.sort(key=lambda entry: source_size(closure(entry))[0])
        if arguments.only:
            entries = entries[:arguments.only]
        for entry in entries:
            label = os.path.relpath(entry, REPOSITORY)
            print(f"proxy {label}", flush=True)
            rows.append(measure_program(arguments, label, "proxy", entry, closure(entry), None, arguments.out))

    # -ftime-trace for the slowest unit at -O2.
    timed = [row for row in rows if row.get("clang") and len(row["clang"]) == 3 and row["clang"][2]["exit"] == 0]
    trace = None
    if timed:
        slowest = max(timed, key=lambda row: row["clang"][2]["wall_seconds"])
        with tempfile.TemporaryDirectory(prefix="tsc-trace-") as directory:
            unit = os.path.join(directory, "unit.c")
            with open(os.path.join(arguments.out, "c", slowest["file"].replace("/", "_") + ".c"), "rb") as source, open(unit, "wb") as handle:
                handle.write(source.read())
            sample = clang_once(arguments.clang, unit, "-O2", directory, trace=True)
            trace = {"file": slowest["file"], "wall_seconds": sample["wall_seconds"], "phases": time_trace_top(directory)}

    clang_version = subprocess.run([arguments.clang, "--version"], capture_output=True, text=True).stdout.split("\n")[0]
    with open("/proc/cpuinfo") as handle:
        cpu = next(line.split(":", 1)[1].strip() for line in handle if line.startswith("model name"))
    with open("/proc/meminfo") as handle:
        memory_kb = int(handle.readline().split()[1])
    box = {"cpu": cpu, "cores": os.cpu_count(), "memory_kb": memory_kb, "clang": clang_version,
           "flags": FLAGS, "kernel": os.uname().release, "load_at_end": loadavg()}
    adamic_commit = subprocess.run(["git", "-C", REPOSITORY, "rev-parse", "HEAD"], capture_output=True, text=True).stdout.strip()
    with open(os.path.join(arguments.out, "rows.json"), "w") as handle:
        json.dump({"adamic_commit": adamic_commit, "box": box, "rows": rows, "time_trace": trace}, handle, indent=1)
        handle.write("\n")
    columns = ["file", "kind", "own_ts_bytes", "own_ts_lines", "closure_files", "closure_ts_bytes", "closure_ts_lines",
               "adamic_exit", "first_reason", "meter_reason", "c_bytes", "c_lines", "c_functions", "units",
               "split_units_if_opted_in", "largest_function", "largest_function_lines"]
    for level in ("O0", "O1", "O2"):
        columns += [f"{level}_seconds", f"{level}_max_rss_kb", f"{level}_exit"]
    with open(os.path.join(arguments.out, "files.csv"), "w", newline="") as handle:
        writer = csv.writer(handle)
        writer.writerow(columns)
        for row in rows:
            values = [row.get(column, "") for column in columns[:17]]
            for level in ("-O0", "-O1", "-O2"):
                sample = next((sample for sample in row.get("clang", []) if sample["level"] == level), None)
                values += [sample["wall_seconds"], sample["max_rss_kb"], sample["exit"]] if sample else ["", "", ""]
            writer.writerow(values)
    return 0


if __name__ == "__main__":
    sys.exit(main())
