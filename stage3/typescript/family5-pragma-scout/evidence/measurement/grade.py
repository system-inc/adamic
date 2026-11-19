"""Pair full-census findings by an aligned source token and the attempting unit."""
from collections import Counter, defaultdict
from difflib import SequenceMatcher
from functools import lru_cache
import json
from pathlib import Path
import re
import sys

scratch = Path(sys.argv[1]).resolve()
output = Path(sys.argv[2])
labels = ("before", "after")
parser = Path("src/compiler/parser.ts")
lines = {
    label: (scratch / label / parser).read_text().split("\n") for label in labels
}
before, after = map(lines.get, labels)
alignment = {}
for op, i, j, k, end in SequenceMatcher(None, before, after, autojunk=False).get_opcodes():
    if op == "equal":
        alignment.update({n + 1: (p + 1, True) for n, p in zip(range(k, end), range(i, j))})
    elif op == "replace" and j - i == end - k:
        for n, p in zip(range(k, end), range(i, j)):
            if SequenceMatcher(None, after[n], before[p], autojunk=False).ratio() >= 0.50:
                alignment[n + 1] = (p + 1, False)


def utf16_column(line, index):
    return len(line[:index].encode("utf-16-le")) // 2 + 1


@lru_cache(maxsize=None)
def column_alignment(new, old):
    left, right = after[new - 1], before[old - 1]
    return {
        utf16_column(left, i + offset): utf16_column(right, j + offset)
        for i, j, size in SequenceMatcher(None, left, right, autojunk=False).get_matching_blocks()
        for offset in range(size)
    }


def location(label, value):
    prefix = str(scratch / label) + "/"
    value = value.removeprefix(prefix)
    match = re.fullmatch(re.escape(str(parser)) + r":(\d+):(\d+)", value)
    if label == "before" or not match:
        return value
    line, col = map(int, match.groups())
    if line in alignment:
        old, exact = alignment[line]
        previous = col if exact else column_alignment(line, old).get(col)
        if previous is not None:
            return f"{parser}:{old}:{previous}"
    return f"after/{value}"


def group(label, finding):
    return tuple(location(label, finding.get(key, "")) for key in ("where", "unit"))


def detail(label, item):
    prefix = str(scratch / label) + "/"
    def normalize_text(value):
        value = value.replace(prefix, "")
        return re.sub(re.escape(str(parser)) + r":\d+:\d+",
                      lambda match: location(label, match.group()), value)

    return tuple(normalize_text(item.get(field, "")) for field in (
        "kind", "phase", "pass", "site_kind", "root_kind", "reason", "text",
    )) + (item.get("depth"),) + tuple(location(label, item.get(field, "")) for field in (
        "site_where", "boundary_where",
    ))


def load(label):
    header, row, trailer = map(json.loads, (scratch / (label + "-latent.jsonl")).read_text().splitlines())
    assert header["latent_mode"] == "full" and header["module_format"] == "CommonJSBundle"
    assert row["file"] == str(scratch / label / parser)
    assert trailer["status"] == "relation_report"
    return header, row, trailer


runs = {label: load(label) for label in labels}
findings = {}
for label in labels:
    groups = defaultdict(list)
    for i, finding in enumerate(runs[label][1]["findings"] or []):
        groups[group(label, finding)].append((i, finding))
    findings[label] = groups

grades = []
def record(grade, key, base=None, head=None):
    grades.append({
        "grade": grade,
        "aligned_where": key[0],
        "aligned_unit": key[1],
        "before": base[0] if base else None,
        "after": head[0] if head else None,
    })

for key in sorted(set(findings["before"]) | set(findings["after"])):
    base = findings["before"].get(key, [])
    head = list(findings["after"].get(key, []))
    unmatched = []
    for old in base:
        j = next((j for j, new in enumerate(head) if
                  detail("before", old[1]) == detail("after", new[1])), None)
        if j is None:
            unmatched.append(old)
        else:
            record("reached", key, old, head.pop(j))
    # Within an aligned site and attempting unit, retain the phase and kind
    # correspondence when several unmatched observations share the location.
    for old in list(unmatched):
        j = next((j for j, new in enumerate(head) if all(
            old[1].get(field) == new[1].get(field) for field in ("phase", "kind")
        )), None)
        if j is not None:
            record("changed", key, old, head.pop(j))
            unmatched.remove(old)
    if len(unmatched) == len(head) == 1:
        record("changed", key, unmatched.pop(), head.pop())
    for old in unmatched:
        record("gone", key, base=old)
    for new in head:
        record("new", key, head=new)

for label in labels:
    indices = sorted(row[label] for row in grades if row[label] is not None)
    assert indices == list(range(len(runs[label][1]["findings"] or [])))
(output / "parser-grades.jsonl").write_text("".join(json.dumps(row) + "\n" for row in grades))
print(dict(Counter(row["grade"] for row in grades)))
for label in labels:
    header, row, trailer = runs[label]
    print(label, len(row["units"]), len(row["findings"]), len(header["diagnostics"]),
          len(trailer["findings"] or []))
units = {
    label: {location(label, unit["where"]): unit for unit in runs[label][1]["units"]}
    for label in labels
}
for key in sorted(set(units["before"]) | set(units["after"])):
    old, new = (units[label].get(key) for label in labels)
    before_status, after_status = (unit and unit["status"] for unit in (old, new))
    if before_status != after_status:
        print("unit status", key, before_status, after_status, (new or old).get("name"))
for row in grades:
    if row["grade"] != "reached":
        print(row["grade"], row["aligned_where"], row["aligned_unit"])
        for label in labels:
            if row[label] is not None:
                f = runs[label][1]["findings"][row[label]]
                print(" ", label, f["where"].removeprefix(str(scratch / label) + "/"),
                      f["phase"], f["kind"], f["reason"], f["text"])
