#!/usr/bin/env python3
"""Retain the five requested codes in the explicitly selected adaptation wave."""
import json
from pathlib import Path
import re
import sys

source = json.loads(Path(sys.argv[1]).read_text())
codes = [2345, 18048, 2532, 2322, 2538]
files = ["core.ts", "utilities.ts", "utilitiesPublic.ts"]
if len(sys.argv) > 3 and sys.argv[3] not in ("", "--wave2", "--wave3", "--scanner-slice"):
    raise SystemExit("usage: census.py <raw-diagnostics> <output> [--wave2|--wave3|--scanner-slice]")
wave2 = len(sys.argv) > 3 and sys.argv[3] == "--wave2"
if wave2:
    files = ["parser.ts", "scanner.ts"] + ["factory/" + name + ".ts" for name in [
        "baseNodeFactory", "emitHelpers", "emitNode", "nodeChildren", "nodeConverters",
        "nodeFactory", "nodeTests", "parenthesizerRules", "utilities", "utilitiesPublic",
    ]]
if len(sys.argv) > 3 and sys.argv[3] == "--wave3":
    files = ["debug.ts", "path.ts"]
if len(sys.argv) > 3 and sys.argv[3] == "--scanner-slice":
    files = ["commandLineParser.ts", "core.ts", "corePublic.ts", "debug.ts",
             "diagnosticInformationMap.generated.ts", "scanner.ts", "types.ts", "utilities.ts"]
pattern = "|".join(re.escape(file) for file in files)
counts = {file: {f"TS{code}": 0 for code in codes} for file in files}
diagnostics = []
for text in source["diagnostics"]:
    match = re.match(r".*/src/compiler/(" + pattern + r"):(\d+):(\d+): error TS(\d+):", text)
    if match and int(match[4]) in codes:
        counts[match[1]][f"TS{match[4]}"] += 1
        diagnostics.append("src/compiler/" + text[match.start(1):])
result = {"roots": source["roots"], "codes": codes, "counts": counts,
          "total": len(diagnostics), "diagnostics": diagnostics}
Path(sys.argv[2]).write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps({"counts": counts, "total": len(diagnostics)}, indent=2))
