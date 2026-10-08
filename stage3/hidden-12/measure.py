"""Audit the assigned binder interval using the preserved complete binder record.

Run with the census pin's hidden.py as the sole argument. That reference is read,
never installed into the delivery compiler. The full project was loaded before
this binder record was emitted; this script makes no whole-corpus total claim.
"""
import gzip
import importlib.util
import json
from pathlib import Path
import sys

here = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("hidden", sys.argv[1])
hidden = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hidden)
with gzip.open(here / "binder-census.jsonl.gz", "rt") as stream:
    rows = [json.loads(line) for line in stream]
with gzip.open(here / "binder-stock.json.gz", "rt") as stream:
    stock = json.load(stream)
record = rows[1]
original = Path(record["file"]).parent
result = hidden.calculate(rows, stock, original)
expected = json.loads((here / "measurement.json").read_text())
left, right = expected["region"]
intersection = hidden.union(
    (max(start, left), min(end, right))
    for start, end in result["files"]["binder.ts"]["hidden_ranges"]
    if start < right and end > left
)
assert [list(span) for span in intersection] == expected["new_hidden_intersection"]
assert stock["binder.ts"]["sha256"] == expected["sha256"]

# Independent byte masks use no union/subtraction helper from hidden.py.
blocked, examined, own = set(), set(), {}
for finding in record["findings"]:
    if finding["kind"] == "Boundary" and finding["where"].rsplit(":", 2)[0] == record["file"]:
        span = set(range(finding["start"], finding["end"]))
        blocked.update(span)
        own.setdefault(finding["unit"], set()).update(span)
skipped = [unit for unit in record["units"] if unit["status"] in ("split_checker_body", "skipped_checker_body")]
for unit in skipped:
    blocked.update(range(unit["body_start"], unit["body_end"]))
catalog = {unit["where"]: unit for unit in stock["binder.ts"]["units"]}
for unit in record["units"]:
    if unit["status"] not in ("attempted", "panic"):
        continue
    span = catalog[unit["where"].replace(str(original) + "/", "")]
    exposed = set(range(span["start"], span["end"])) - own.get(unit["where"], set())
    for child in skipped:
        if span["start"] <= child["body_start"] and child["body_end"] <= span["end"]:
            exposed.difference_update(range(child["body_start"], child["body_end"]))
    examined.update(exposed)
mask = (blocked - examined) & set(range(left, right))
assert len(mask) == hidden.size(intersection) == expected["new_hidden_bytes"] == right - left
# Non-binder records cannot expose binder bytes; dependency events only add
# blocks/cuts. A lower bound equal to the interval's size is therefore exact.
print(f"PASS: hidden intersection [{left},{right}) = {len(mask)} bytes; revealed {expected['revealed_bytes']}")
