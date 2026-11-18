"""Compare findings in the two completed tsc entry streams."""
from collections import Counter
import gzip
import json
from pathlib import Path
import re

evidence = Path(__file__).resolve().parent
phrases = ("polymorphic recursion", "instantiates itself deeper")
position = re.compile(r"(\.(?:ts|tsx|a))(?::\d+:\d+|\(\d+\s*,\s*\d+\))")


def load(label):
    path = evidence / label / "solo-entry" / "latent.jsonl.gz"
    with gzip.open(path, "rt") as stream:
        raw = stream.read()
    header, *files, trailer = map(json.loads, raw.splitlines())
    assert header["status"] == "measurement" and not header["checker_rejected"]
    assert header["module_format"] == "CommonJSBundle"
    assert header["latent_mode"] == "first-error"
    (entry,) = header["roots"]
    assert entry.endswith("/src/tsc/tsc.ts")
    assert trailer["status"] == "relation_report"
    assert sorted(row["file"] for row in files) == sorted(header["sources"])
    findings = [finding for row in [*files, trailer] for finding in row.get("findings") or []]
    return raw, entry.removesuffix("/src/tsc/tsc.ts"), findings


def key(finding, tree):
    return tuple(
        position.sub(r"\1", finding[field].replace(tree + "/", ""))
        for field in ("kind", "phase", "reason", "text")
    )


_, base_tree, base = load("base")
head_raw, head_tree, head = load("revised-head")
assert base_tree == head_tree
before = Counter(key(finding, base_tree) for finding in base)
after = Counter(key(finding, head_tree) for finding in head)
# The shared tree makes exact source locations usable as a second check for
# later findings reached in units that already reported other findings.
before_units = Counter((finding["unit"], key(finding, base_tree)) for finding in base)
after_units = Counter((finding["unit"], key(finding, head_tree)) for finding in head)
assert before == after
assert before_units == after_units
assert base == head
assert not any(phrase in head_raw for phrase in phrases)
print(json.dumps({
    "base_findings": len(base),
    "head_findings": len(head),
    "tail_findings_per_side": sum(finding["phase"] == "tail" for finding in head),
    "new": sum((after - before).values()),
    "gone": sum((before - after).values()),
    "changed": 0,
    "newly_reached": sum((after_units - before_units).values()),
    "head_phrase_hits": {phrase: head_raw.count(phrase) for phrase in phrases},
    "exact_finding_records_equal": base == head,
}, indent=2))
