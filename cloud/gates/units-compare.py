#!/usr/bin/env python3
"""Compare two latent census passes unit by unit, base against head, for cloud/gates/unit-census-pair.sh.

Usage: units-compare.py <base prefix> <head prefix> <base tree> <head tree> [--message <file>]

Each prefix names a pass's census rows, <prefix>.jsonl. A file row carries `units` (each with `where`, `name`, `kind`)
and `findings` (each with `unit`, the where of its unit, and `phase`, `kind`, `where`, `reason`). A unit is keyed by
its file, name and kind and its ordinal among same-named units in that file, never by its line, since an edit shifts
lines; that's why this isn't stage3/census/pair/pair.py itself, which needs the same adapted bytes on both sides and so
refuses an adaptation's pair. Its rule is pair.py's: a unit's first stop is its first finding in phase lowering,
deferred or prepass, and a unit newly stopping has no first stop in the base and one in the head. A unit only the head
holds counts as newly stopping when it stops, fail-closed (Adamic's call on #v3q9ayd; Correctness owns the census rules
and may amend it); a unit only the base holds, or one only the head holds that doesn't stop, is listed by name and fails
nothing. Seats' compare on
J-f910xec, fixed after Correctness caught the version that read the times file and printed zeros (J-ffhyf3v), so it
fails when either side has fewer than 1,000 units (UNITS_COMPARE_MINIMUM): a stream it didn't read isn't 0 new.

A unit newly stopping is allowed only by a ruling line in the head's commit message (--message), alone on its line and
in exactly this form, its file, its name and the ruling's task:

    allowed: src/compiler/checker.ts isArrayType #abc1234

Anything else allows nothing: a mention, a denial ("is not allowed by #abc1234"), or a ruling line with more words.
A line names a file and a name, never an instance, so it covers a unit only when that name is one unit in its file at
the base and at the head; a name held twice, or only on one side, allows nothing by a line.
A unit name whose count of instances differs between base and head can't be matched instance to instance (a function
removed and a quiet namesake now stopping read alike), so every instance of it that stops in the head counts as newly
stopping, and the name is listed. Exit 0 with none newly stopping that isn't allowed, 1 with one, 2 when a side
wasn't read.
"""

import argparse
import collections
import json
import os
from pathlib import Path
import re
import sys

position = re.compile(r"(\.[cm]?ts|\.a)(?::\d+(?::\d+)?|\(\d+,\s*\d+\))")
stopPhases = ("lowering", "deferred", "prepass")


def rows(prefix):
    read = []
    path = Path(prefix + ".jsonl")
    if path.exists():
        for line in path.read_text().splitlines():
            try:
                read.append(json.loads(line))
            except json.JSONDecodeError:
                break
    return read


def strip(text, tree):
    return str(text).replace(tree.rstrip("/") + "/", "")


def side(census, tree):
    """The side's unit keys and each stopping unit's first stop."""
    keys, first = {}, {}
    for row in census:
        file = strip(row.get("file", ""), tree)
        seen = collections.Counter()
        for unit in row.get("units") or []:
            name = (file, unit.get("name", ""), unit.get("kind", ""))
            seen[name] += 1
            keys[strip(unit.get("where", ""), tree)] = name + (seen[name],)
        for finding in row.get("findings") or []:
            if finding.get("phase") not in stopPhases:
                continue
            where = strip(finding.get("unit", ""), tree)
            key = keys.get(where, ("unlisted", where, "", 0))
            first.setdefault(key, (finding.get("kind"), strip(finding.get("where", ""), tree),
                                   position.sub(r"\1", strip(finding.get("reason", ""), tree)), finding.get("phase")))
    return set(keys.values()), first


def compare(base, head, base_tree, head_tree):
    base_units, base_first = side(base, base_tree)
    head_units, head_first = side(head, head_tree)
    both = base_units & head_units
    # A name both sides hold, a different number of times: its instances can't be matched by ordinal, so every instance
    # of it stopping in the head is newly stopping, fail-closed. A name one side alone holds is the head-only rule's.
    base_names = collections.Counter(unit[:3] for unit in base_units)
    head_names = collections.Counter(unit[:3] for unit in head_units)
    ambiguous = sorted(name for name in base_names.keys() & head_names.keys() if base_names[name] != head_names[name])
    ambiguous_names = set(ambiguous)
    unmatched = {unit for unit in head_units if unit[:3] in ambiguous_names and unit in head_first}
    # A ruling line names a unit by its file and name alone, so it covers one only when that name is one unit in its file
    # on each side, whatever its kind: a name held twice, or held on one side alone, is no single instance to rule on
    base_counts = collections.Counter(unit[:2] for unit in base_units)
    head_counts = collections.Counter(unit[:2] for unit in head_units)
    return {
        "ambiguous": ambiguous,
        "unique": {name for name in head_counts if head_counts[name] == 1 and base_counts[name] == 1},
        "base_units": base_units, "head_units": head_units, "base_first": base_first, "head_first": head_first,
        # Every unit the head holds, not only those both sides hold: one only the head holds that stops is newly
        # stopping, since no base stands beside it to say it stopped before
        "newly": sorted({u for u in head_units if u in head_first and u not in base_first} | unmatched),
        "base_only": sorted(base_units - head_units),
        "head_only_quiet": sorted(u for u in head_units - base_units if u not in head_first),
        "cleared": sorted(u for u in both if u in base_first and u not in head_first),
        "moved": sorted(u for u in both if u in base_first and u in head_first and base_first[u][2] != head_first[u][2]),
    }


RulingLine = re.compile(r"allowed: (\S+) (\S+) #[0-9a-z]{6,7}")


def allowed(unit, message, unique):
    """The commit message's ruling line for the unit, `allowed: <file> <name> #<task>` alone on its line, or None. A line
    covers the unit only when its file and name are one unit on each side (compare's unique): otherwise it allows
    nothing, since it can't say which instance it rules on (Seats' default on #v3q9ayd, Correctness's to amend)."""
    if unit[:2] not in unique:
        return None
    for line in message.splitlines():
        found = RulingLine.fullmatch(line.strip())
        if found and found.group(1) == unit[0] and found.group(2) == unit[1]:
            return line.strip()
    return None


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("base_prefix")
    parser.add_argument("head_prefix")
    parser.add_argument("base_tree")
    parser.add_argument("head_tree")
    parser.add_argument("--message", help="the head's commit message, whose lines may allow a newly stopping unit")
    arguments = parser.parse_args()
    result = compare(rows(arguments.base_prefix), rows(arguments.head_prefix), arguments.base_tree, arguments.head_tree)
    minimum = int(os.environ.get("UNITS_COMPARE_MINIMUM", "1000"))
    for label in ("base", "head"):
        units = result[label + "_units"]
        if len(units) < minimum:
            print(f"FAIL units-compare: {label} has {len(units)} units, under {minimum}: the stream was not read as expected")
            return 2
    base_units, head_units = result["base_units"], result["head_units"]
    print(f"  units: base {len(base_units)}, head {len(head_units)}, both {len(base_units & head_units)}; "
          f"only base {len(base_units - head_units)}, only head {len(head_units - base_units)}; "
          f"stopping: base {len(result['base_first'])}, head {len(result['head_first'])}")
    print(f"  unit names held a different number of times on each side, every stopping instance newly stopping: "
          f"{len(result['ambiguous'])}")
    for name in result["ambiguous"]:
        print(f"    ambiguous: {name}")
    print(f"  units only in the base, failing nothing: {len(result['base_only'])}")
    for unit in result["base_only"]:
        print(f"    only in base: {unit}  first stop: {result['base_first'].get(unit)}")
    print(f"  units only in the head with no stop, failing nothing: {len(result['head_only_quiet'])}")
    for unit in result["head_only_quiet"]:
        print(f"    only in head: {unit}")
    message = Path(arguments.message).read_text() if arguments.message else ""
    refused = 0
    print(f"  units newly stopping: {len(result['newly'])}")
    for unit in result["newly"]:
        ruling = allowed(unit, message, result["unique"])
        only = "" if unit in base_units else " (only in the head)"
        print(f"    NEW {unit}{only}\n        head: {result['head_first'][unit]}")
        print(f"        allowed by the commit message: {ruling}" if ruling else
              "        not allowed: no ruling line for it, or its name isn't one unit in its file on each side")
        refused += ruling is None
    print(f"  units cleared: {len(result['cleared'])}")
    for unit in result["cleared"]:
        print(f"    cleared {unit}\n        base: {result['base_first'][unit]}")
    print(f"  units whose first stop moved: {len(result['moved'])}")
    for unit in result["moved"]:
        print(f"    moved {unit}\n        base: {result['base_first'][unit]}\n        head: {result['head_first'][unit]}")
    print(f"  newly stopping not allowed: {refused}")
    return 1 if refused else 0


if __name__ == "__main__":
    sys.exit(main())
