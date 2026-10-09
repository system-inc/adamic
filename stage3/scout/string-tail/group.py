#!/usr/bin/env python3
"""Group the test262 runner's first refusals, without claiming later blockers vanish."""
import collections
import json
import sys


def owner(reason):
    if "named capturing groups" in reason:
        return "regex"
    if any("(" + name + " would" in reason for name in ("search", "match", "replace", "split")):
        return "regex/compiler method identity"
    if "ToPrimitive" in reason or "Object.defineProperty" in reason or "isPrototypeOf" in reason:
        return "objects/ruling"
    if "try around repeat" in reason or "localeCompare (Node" in reason or "inherited library member toLocale" in reason:
        return "library"
    return "compiler"


def grouped(document):
    counts = collections.Counter()
    reasons = []
    for report in document["filters"]:
        for entry in report["refusalReasons"]:
            assigned = owner(entry["reason"])
            counts[assigned] += entry["count"]
            reasons.append(dict(entry, owner=assigned))
    return {"owners": dict(counts), "reasons": reasons}


if __name__ == "__main__":
    with open(sys.argv[1]) as source:
        print(json.dumps(grouped(json.load(source)), indent=2))
