#!/usr/bin/env python3
"""Reconcile instruction-address Callgrind branch profiles without double-counting edges."""
import argparse
import collections
import json
import re
from pathlib import Path


def read_profile(path):
    symbols = {"fn": {}, "fl": {}}
    function = source = "???"
    positions = [0, 0]
    pending = False
    own = collections.defaultdict(lambda: [0] * len(events))
    instructions = collections.defaultdict(lambda: [0] * len(events))
    events = summary = footer = None
    taken = collections.Counter()
    for row in path.read_text().splitlines():
        if row.startswith("events:"):
            events = row.split()[1:]
        elif row.startswith("summary:"):
            summary = list(map(int, row.split()[1:]))
        elif row.startswith("totals:"):
            footer = list(map(int, row.split()[1:]))
        elif row.startswith(("fn=", "fl=", "fi=", "fe=", "cfn=", "cfl=", "cfi=", "cfe=", "jfi=", "jfl=", "jfn=")):
            key, value = row.split("=", 1)
            category = "fn" if key.endswith("fn") else "fl"
            match = re.fullmatch(r"\((\d+)\)(?: (.*))?", value)
            if match:
                ident, name = match.groups()
                if name is not None:
                    symbols[category][ident] = name
                value = symbols[category][ident]
            if key == "fn":
                function = value
            elif key in ("fl", "fi", "fe"):
                source = value
        elif row.startswith("jcnd="):
            count = int(row.split()[0][5:].split("/")[0])
            taken[function, source, positions[0], positions[1]] += count
        elif row.startswith("calls="):
            pending = True
        elif row and row[0] in "0123456789+-*":
            fields = row.split()
            for i, value in enumerate(fields[:2]):
                if value != "*":
                    if value[0] in "+-":
                        positions[i] += int(value, 0) if "x" in value else int(value)
                    else:
                        positions[i] = int(value, 0) if "x" in value else int(value)
            costs = list(map(int, fields[2:]))
            costs += [0] * (len(events) - len(costs))
            if pending:
                pending = False
                continue
            for i, value in enumerate(costs):
                own[function][i] += value
                instructions[function, source, positions[0], positions[1]][i] += value
    if not events or events[0] != "Ir" or "I1mr" not in events or "Bc" not in events:
        raise ValueError("unexpected event order")
    sums = [sum(v[i] for v in own.values()) for i in range(len(events))]
    if sums != footer:
        raise ValueError(f"self/footer mismatch: {sums} / {footer}")
    header_delta = [a-b for a,b in zip(summary,footer)]
    if header_delta not in ([0]*len(events), [2]+[0]*(len(events)-1)):
        raise ValueError(f"unexpected header/footer discrepancy: {header_delta}")
    return dict(events=events, totals=sums, header=summary, header_delta=header_delta,
                functions=[dict(function=fn, costs=v) for fn, v in
                           sorted(own.items(), key=lambda x: x[1][2] + x[1][4], reverse=True)],
                instructions=[dict(function=fn, source=source, address=hex(addr), line=line, costs=v, taken=taken[fn, source, addr, line])
                              for (fn, source, addr, line), v in instructions.items() if fn == "adamic_object_free_children" or v[events.index("Bc")] or v[events.index("Bi")]])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("profile", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    result = read_profile(args.profile)
    args.output.write_text(json.dumps(result, indent=2) + "\n")
    print(dict(zip(result["events"], result["totals"])))
    for row in result["functions"][:12]:
        print(row)
    for row in result["instructions"]:
        if row["function"] == "adamic_release":
            print(row)


if __name__ == "__main__":
    main()
