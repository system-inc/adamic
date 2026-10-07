"""Audit recorded census and default oracle observations without running them."""
import collections
import json
import gzip
from pathlib import Path
import re
import sys

before_path, after_path, oracle_path = map(Path, sys.argv[1:])
def records(path):
    text = gzip.decompress(path.read_bytes()).decode() if path.suffix == ".gz" else path.read_text()
    return [json.loads(line) for line in text.splitlines()]


before = records(before_path)
after = records(after_path)
oracle = json.loads((oracle_path / "report.json").read_text())


def check(before, after, oracle, baseline):
    assert len(before[-1]["roots"]) == len(after[-1]["roots"]) == 78
    assert len(before) == len(after)
    diagnostics = before[-1].get("diagnostics", [])
    counts = collections.Counter(int(code) for text in diagnostics for code in re.findall(r"error TS(\d+):", text))
    assert counts == {2345: 15, 18048: 6, 2322: 4, 2532: 4}, counts
    assert len(diagnostics) == 29
    assert not after[-1].get("diagnostics"), after[-1]
    source_entries = [entry for entry in after[:-1] if entry["roots"][0].endswith(".ts")]
    assert len(source_entries) == 78
    assert all(not entry.get("diagnostics") for entry in source_entries)
    assert oracle["status"] == "pass"
    assert oracle["runners"] == "all" and oracle["tests"] is None and oracle["workers"] == 4
    assert all(phase["exit"] == 0 for phase in oracle["phases"].values())
    assert oracle["counts"]["passing"] > 0 and oracle["counts"]["failing"] == 0
    assert not oracle["baseline_diffs"]
    assert baseline == b""


baseline = (oracle_path / "baseline.diff").read_bytes()
check(before, after, oracle, baseline)
print("PASS: upstream-config census 29 -> 0, every one of 78 source entries diagnostic-free")
print("PASS: unfiltered default oracle, all phase exits zero, baseline.diff 0 bytes")
for name, mutate in [
    ("retained diagnostic", lambda: check(before, [*after[:-1], dict(after[-1], diagnostics=[before[-1]["diagnostics"][0]])], oracle, baseline)),
    ("baseline difference", lambda: check(before, after, oracle, b"mutant baseline difference\n")),
    ("filtered oracle", lambda: check(before, after, dict(oracle, tests="semver"), baseline)),
]:
    try:
        mutate()
    except AssertionError:
        print(f"PASS: {name} record mutant caught by its evidence assertion")
    else:
        raise AssertionError(f"surviving evidence mutant: {name}")
print(json.dumps({"before": 29, "after": 0, "roots": 78, "oracle_counts": oracle["counts"], "baseline_bytes": len(baseline)}, indent=2))
