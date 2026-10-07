#!/usr/bin/env python3
"""Retain the requested codes on exactly the reviewed files."""
import json
from pathlib import Path
import re
import sys
source = json.loads(Path(sys.argv[1]).read_text())
codes = [2345, 18048, 2532, 2322, 2538]
files = json.loads(Path(__file__).with_name("files.json").read_text())
counts = {file: {f"TS{code}": 0 for code in codes} for file in files}
diagnostics = []
for text in source["diagnostics"]:
    match = re.match(r".*/src/compiler/(.*?):(\d+):(\d+): error TS(\d+):", text)
    if match and match[1] in counts and int(match[4]) in codes:
        counts[match[1]][f"TS{match[4]}"] += 1
        diagnostics.append("src/compiler/" + text[match.start(1):])
result = {"roots": source["roots"], "codes": codes, "counts": counts,
          "total": len(diagnostics), "diagnostics": diagnostics}
Path(sys.argv[2]).write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps({"counts": counts, "total": len(diagnostics)}, indent=2))
