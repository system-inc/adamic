#!/usr/bin/env python3
"""List every count/corpus inventory or answer change, without editing either output."""
import json
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
report = Path(__file__).resolve().parent

def prior(ref, path):
    return subprocess.check_output(["git", "show", f"{ref}:{path}"], cwd=root, timeout=30).decode()

def counts(text):
    result = {}
    occurrences = {}
    for line in text.splitlines():
        cells = [cell.strip() for cell in line.split("|")]
        if len(cells) > 3 and cells[1].startswith("internal/"):
            fixture = cells[1]
            occurrences[fixture] = occurrences.get(fixture, 0) + 1
            key = fixture if occurrences[fixture] == 1 else f"{fixture} [occurrence {occurrences[fixture]}]"
            result[key] = line
    return result

path = "internal/oracle/counts.md"
before = counts(prior("285d7ef9", path))
after = counts((root / path).read_text())
lines = []
for key in sorted(before.keys() | after.keys()):
    if key not in before:
        lines.extend(["ADDED " + key, after[key]])
    elif key not in after:
        lines.extend(["REMOVED " + key, before[key]])
    elif before[key] != after[key]:
        lines.extend(["MOVED " + key, "before: " + before[key], "after:  " + after[key]])
(report / "counts-rows.txt").write_text("\n".join(lines) + "\n")
removed = before.keys() - after.keys()
print(f"Counts: {len(before)} -> {len(after)}; added={len(after.keys()-before.keys())}, moved={sum(before[k]!=after[k] for k in before.keys() & after.keys())}, removed={len(removed)}")
if removed:
    raise RuntimeError("existing Linux rows removed: " + ", ".join(sorted(removed)))

path = "internal/lower/testdata/cycles.json"
after = json.loads((root / path).read_text())
for ref, name in [("1e6140fe", "corpus-rows.txt"), ("03ed6ebf", "corpus-regeneration-rows.txt")]:
    before = json.loads(prior(ref, path))
    lines = []
    for key in sorted(before.keys() | after.keys()):
        if key not in before:
            lines.append("ADDED " + key)
        elif key not in after:
            lines.append("REMOVED " + key)
        elif before[key] != after[key]:
            fields = sorted(k for k in before[key].keys() | after[key].keys() if before[key].get(k) != after[key].get(k))
            lines.append("MOVED " + key + " fields=" + ",".join(fields) + " status=" + before[key]["status"] + "->" + after[key]["status"])
    (report / name).write_text("\n".join(lines) + "\n")
    print(f"Corpus vs {ref}: {len(before)} -> {len(after)}; added={len(after.keys()-before.keys())}, moved={sum(before[k]!=after[k] for k in before.keys() & after.keys())}, removed={len(before.keys()-after.keys())}")
