#!/usr/bin/env python3
"""Seats: print every unit's first stop from one census pass, by pair.py's rule.

Usage: firststops.py <prefix> <tree> <pass>
Reads <prefix>.jsonl. A unit's first stop is its first finding in phase lowering, deferred or prepass. Prints one
tab-separated line per stopping unit: FS, pass, file, unit name, unit kind, ordinal, finding kind, where, reason.
Exits 1 when it reads fewer than 1,000 units, so an unread stream never passes as an empty table.
"""
import collections, json, sys
from pathlib import Path

prefix, tree, pass_name = sys.argv[1:4]
root = tree.rstrip("/") + "/"
units = 0
rows = []
for line in Path(prefix + ".jsonl").read_text().splitlines():
    try:
        rows.append(json.loads(line))
    except json.JSONDecodeError:
        break
for row in rows:
    file = str(row.get("file", "")).replace(root, "")
    keys, seen, first = {}, collections.Counter(), {}
    for unit in row.get("units") or []:
        name = (unit.get("name", ""), unit.get("kind", ""))
        seen[name] += 1
        keys[str(unit.get("where", "")).replace(root, "")] = name + (seen[name],)
        units += 1
    for finding in row.get("findings") or []:
        if finding.get("phase") not in ("lowering", "deferred", "prepass"):
            continue
        key = keys.get(str(finding.get("unit", "")).replace(root, ""))
        if key is None:
            continue
        first.setdefault(key, finding)
    for key, finding in first.items():
        print("\t".join(["FS", pass_name, file, key[0], key[1], str(key[2]), str(finding.get("kind")),
                         str(finding.get("where", "")).replace(root, ""), str(finding.get("reason", "")).replace(root, "").replace("\t", " ").replace("\n", " ")]))
print(f"UNITS\t{pass_name}\t{units}")
if units < 1000:
    print(f"FAIL firststops: {units} units read in {pass_name}, under 1,000")
    sys.exit(1)
