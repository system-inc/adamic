"""Pair findings across the first-error streams in baseline source coordinates."""

from collections import Counter, defaultdict, deque
from difflib import SequenceMatcher
from functools import cache
import gzip
import json
from pathlib import Path
import re
import sys


root = Path("/tmp/adamic-g02")
evidence = Path(__file__).resolve().parent
target = (
    "optional property getPositionOfLineAndCharacter in SourceFileLike "
    "absent from structural source SourceFile, which can hide fields"
)
position = re.compile(re.escape(str(root)) + r"/(?:base|edit)/(src/[^:\s]+):(\d+):(\d+)")


@cache
def line_map(file):
    before = (root / "base" / file).read_text().splitlines()
    after = (root / "edit" / file).read_text().splitlines()
    blocks = SequenceMatcher(None, before, after, autojunk=False).get_opcodes()
    return before, after, blocks


def normalize(value, side):
    def replace(match):
        file, line, column = match.groups()
        if side == "edit" and file in edited_files:
            before, after, blocks = line_map(file)
            edited_index = int(line) - 1
            for kind, start, end, edit_start, edit_end in blocks:
                if not edit_start <= edited_index < edit_end:
                    continue
                index = start + min(edited_index - edit_start, max(end - start - 1, 0))
                line = str(index + 1)
                if kind == "replace":
                    offset = int(column) - 1
                    for _, old_lo, old_hi, new_lo, new_hi in SequenceMatcher(
                        None, before[index], after[edited_index], autojunk=False
                    ).get_opcodes():
                        if new_lo <= offset < new_hi:
                            column = str(old_lo + min(offset - new_lo, max(old_hi - old_lo - 1, 0)) + 1)
                            break
                break
        return f"{file}:{line}:{column}"

    return position.sub(replace, value).replace(str(root / side) + "/", "")


def read(side):
    out = evidence / "entry" / side
    metadata = json.loads((out / "status.json").read_text())
    assert metadata["exit_code"] == 0, metadata
    with gzip.open(out / "latent.jsonl.gz", "rt") as stream:
        rows = [json.loads(line) for line in stream]
    assert rows[0]["checker_rejected"] is False and not rows[0]["diagnostics"]
    assert rows[0]["module_format"] == "CommonJSBundle" and rows[0]["latent_mode"] == "first-error"
    assert rows[-1]["status"] == "relation_report"
    return [
        {key: normalize(value, side) if isinstance(value, str) else value for key, value in finding.items()}
        for row in rows
        for finding in row.get("findings", [])
    ]


def key(f):
    return tuple(f.get(field) for field in ("unit", "where", "kind", "phase", "reason", "text", "root_kind", "depth"))


def site(f):
    return f.get("unit"), f.get("where"), f.get("phase")


def order(f):
    filename, line, column = f["where"].rsplit(":", 2)
    return filename, int(line), int(column)


edited_files = {
    "src/compiler/scanner.ts", "src/compiler/utilities.ts", "src/compiler/factory/nodeChildren.ts",
}
before = read("base")
after = read("edit")
unchanged = 0
after_by_key = defaultdict(deque)
for i, f in enumerate(after):
    after_by_key[key(f)].append(i)
unmatched_before = []
unmatched_after = set(range(len(after)))
for f in before:
    matches = after_by_key[key(f)]
    if matches:
        unmatched_after.remove(matches.popleft())
        unchanged += 1
    else:
        unmatched_before.append(f)

changes = []
left_before = []
for f in unmatched_before:
    corresponding = next((i for i in sorted(unmatched_after) if site(after[i]) == site(f)), None)
    if corresponding is None:
        left_before.append(f)
    else:
        unmatched_after.remove(corresponding)
        changes.append({"grade": "changed", "before": f, "after": after[corresponding]})

blocked = defaultdict(list)
for f in left_before:
    if f.get("reason") == target and f.get("unit"):
        blocked[f["unit"]].append(f)

for i in sorted(unmatched_after):
    f = after[i]
    old_stops = blocked.get(f.get("unit"), [])
    first_later = next((
        old for old in old_stops if
        f["phase"] in {"lowering", "deferred", "registration", "prepass", "tail"} or
        f["phase"] == old["phase"] and order(f)[0] == order(old)[0] and order(f) > order(old)
    ), None)
    # A later source location is insufficient when the stream already scanned the same
    # phase of that unit. In that case retain the finding as new.
    if first_later is not None and f["phase"] != first_later["phase"]:
        changes.append({"grade": "reached", "after": f, "prior_stop": first_later})
    else:
        changes.append({"grade": "new", "after": f})

changes.extend({"grade": "gone", "before": f} for f in left_before)
summary = {
    "base_findings": len(before),
    "edit_findings": len(after),
    "unchanged": unchanged,
    "target_base_sites": sorted({f["where"] for f in before if f["reason"] == target}),
    "target_edit_sites": sorted({f["where"] for f in after if f["reason"] == target}),
    "grades": dict(Counter(change["grade"] for change in changes)),
}
assert unchanged + sum("before" in change for change in changes) == len(before)
assert unchanged + sum("after" in change for change in changes) == len(after)
(evidence / "comparison.json").write_text(json.dumps({"summary": summary, "changes": changes}, indent=2) + "\n")
print(json.dumps(summary, indent=2))
for change in changes:
    sides = [change[side] for side in ("before", "after") if side in change]
    print(change["grade"], " -> ".join(f"{f['where']} {f['phase']} {f['reason']}" for f in sides))
fails = summary["target_edit_sites"] or any(
    change["grade"] in {"new", "changed"} or
    change["grade"] == "gone" and change["before"]["reason"] != target
    for change in changes
)
sys.exit(bool(fails))
