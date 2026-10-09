#!/usr/bin/env python3
"""Compare actual clean-main and extended full oracle/lane observations."""
import hashlib, json, sys
from pathlib import Path
unit = Path(__file__).resolve().parent
before, after = map(Path, sys.argv[1:3])
b, a = [json.loads((p / "oracle/report.json").read_text()) for p in (before, after)]
bl, al = [json.loads((p / "report.json").read_text()) for p in (before, after)]
fields = ("status", "counts", "baseline_diffs", "node", "runners", "tests", "workers")
for name in fields:
    assert b[name] == a[name], (name, b[name], a[name])
assert {k:v["exit"] for k,v in b["phases"].items()} == {k:v["exit"] for k,v in a["phases"].items()}
for name in ("status", "counts", "failed_tests", "baseline_diffs", "api", "platform", "verdict"):
    assert bl[name] == al[name], (name, bl[name], al[name])
assert bl["status"] == al["status"] == "pass" and not bl["errors"] and not al["errors"]
bd, ad = [(p / "oracle/baseline.diff").read_bytes() for p in (before, after)]
assert bd == ad, "upstream baseline diff bytes changed"
report = {"oracle": {k:a[k] for k in fields}, "phaseExits": {k:v["exit"] for k,v in a["phases"].items()},
          "baselineDiffIdentical": True, "baselineDiffSha256": hashlib.sha256(ad).hexdigest(),
          "lane": {k:al[k] for k in ("status", "counts", "failed_tests", "api", "platform", "verdict")},
          "controlMethod": "independent archived-main full apply and full oracle; unchanged lane checker replays those measured results",
          "extensionMethod": "independently measured full apply; isolated unchanged full oracle; unchanged lane checker replay"}
(unit / "gate-comparison.json").write_text(json.dumps(report, indent=2) + "\n")
print(json.dumps(report, indent=2))
