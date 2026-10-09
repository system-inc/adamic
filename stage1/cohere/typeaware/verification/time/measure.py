#!/usr/bin/env python3
"""Summarize verbose typeaware phase logs and optional clang child records."""
import argparse
import collections
import json
import pathlib
import re

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("log", type=pathlib.Path)
parser.add_argument("--clang", type=pathlib.Path)
args = parser.parse_args()
text = args.log.read_text()
phases = collections.defaultdict(list)
for phase, name, seconds in re.findall(r"phase (\w+) (\S+) ([0-9.]+)s", text):
    if phase == "command":
        if "checker" in name:
            phase = "checker_archive"
        elif name == "stage0":
            phase = "stage0_build"
        elif name.endswith("-build"):
            phase = "other_builds"
        elif name.endswith("-compiler-node-test"):
            phase = "compiler_mutant_tests"
        else:
            phase = "runs"
    phases[phase].append({"name": name, "seconds": float(seconds)})
result = {
    "phases": {
        phase: {"count": len(rows), "sum_seconds": sum(row["seconds"] for row in rows), "invocations": rows}
        for phase, rows in sorted(phases.items())
    },
    "outcome": "timeout" if "panic: test timed out" in text else ("pass" if "\nPASS\n" in text else "incomplete"),
    "package_wall_seconds": re.findall(r"(?:ok|FAIL)\s+\S+\s+([0-9.]+)s", text),
    "tests": re.findall(r"--- (PASS|FAIL): (\S+) \(([0-9.]+)s\)", text),
}
if args.clang:
    result["clang_children"] = [json.loads(path.read_text()) for path in sorted(args.clang.glob("*.json"))]
print(json.dumps(result, indent=2))
