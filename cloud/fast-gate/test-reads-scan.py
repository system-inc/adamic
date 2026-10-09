#!/usr/bin/env python3
"""Scan every _test.go for path literals that reach another package's test inputs (its _test.go files, a testdata
dir, or review/), resolved against the test's package directory, and print cloud/fast-gate/test-reads.json."""
import json, os, re, subprocess, sys
root = sys.argv[1]
files = subprocess.run(["git", "-C", root, "ls-files", "*_test.go"], capture_output=True, text=True, check=True).stdout.split()
literal = re.compile(r'"((?:\.\./)+[^"\s]*|review/[^"\s]*)"')
reads = {}
for path in files:
    if path.startswith("cohere/"):
        continue
    package = os.path.dirname(path)
    with open(os.path.join(root, path), errors="replace") as handle:
        source = handle.read()
    for match in literal.findall(source):
        target = os.path.normpath(os.path.join(package, match)) if match.startswith("..") else match
        if target.startswith("..") or target == ".":
            continue
        parts = target.split("/")
        if "testdata" in parts:
            target = "/".join(parts[: parts.index("testdata") + 1])
        elif target.startswith("review/"):
            target = "/".join(parts[:2])
        elif target.endswith("_test.go") or (os.path.isdir(os.path.join(root, target)) and any(name.endswith("_test.go") for name in os.listdir(os.path.join(root, target)))):
            target = target if target.endswith("_test.go") else target
        else:
            continue
        if target == package or target.startswith(package + "/testdata"):
            continue
        reads.setdefault(package, set()).add(target)
reads["cmd/adamic-gate"] = {"."}
print(json.dumps({key: sorted(value) for key, value in sorted(reads.items())}, indent=2))
