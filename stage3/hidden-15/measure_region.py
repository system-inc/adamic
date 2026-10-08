"""Calculate hidden-15 against the pin's stock catalogue and hidden arithmetic."""
import argparse, gzip, hashlib, importlib.util, json
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("pin_tree", type=Path)
parser.add_argument("compiler", type=Path)
parser.add_argument("census", type=Path)
parser.add_argument("output", type=Path)
args = parser.parse_args()
root = args.pin_tree / "stage3/census/hidden"
spec = importlib.util.spec_from_file_location("hidden", root / "hidden.py")
hidden = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hidden)
stock = hidden.read_json(root / "evidence/stock.json.gz")
old = hidden.read_json(root / "RESULT.json")
for name, metadata in old["files"].items():
    actual = (args.compiler / name).read_bytes()
    assert len(actual) == metadata["bytes"]
    assert hashlib.sha256(actual).hexdigest() == metadata["sha256"], name
rows = hidden.read_rows(args.census)
assert len(rows) == 2 and rows[1]["file"] == str(args.compiler / "program.ts")
new = hidden.calculate(rows, {"program.ts": stock["program.ts"]}, args.compiler)
start, end = 117206, 123466
def intersect(spans):
    return hidden.union((max(left, start), min(right, end)) for left, right in spans
                        if left < end and right > start)
before = intersect(old["files"]["program.ts"]["hidden_ranges"])
after = intersect(new["files"]["program.ts"]["hidden_ranges"])
result = {"interval": [start, end], "old_hidden_intersections": before,
          "new_hidden_intersections": after, "old_hidden_bytes": hidden.size(before),
          "new_hidden_bytes": hidden.size(after),
          "revealed_bytes": hidden.size(before) - hidden.size(after)}
assert result["old_hidden_bytes"] == 6260
# Independently catalogue every declaration that could expose bytes in this region.
overlapping = [unit for unit in stock["program.ts"]["units"]
               if unit["start"] < end and unit["end"] > start]
assert [unit["where"] for unit in overlapping] == ["program.ts:1515:1", "program.ts:2345:5"]
outer, attempt = [next(unit for unit in rows[1]["units"]
                      if unit["where"].endswith("/" + known["where"]))
                  for known in overlapping]
assert outer["status"] == "split_checker_body"
assert attempt["status"] == "attempted"
assert any(finding["kind"] == "Boundary" and finding.get("start") == start
           and finding.get("end") == end and finding["unit"] == attempt["where"]
           for finding in rows[1]["findings"])
# No independently attempted nested declaration can expose a subrange. The
# region stays completely blocked even if another file adds dependency skips.
assert result["new_hidden_bytes"] == 6260
args.output.write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps(result, indent=2))
