#!/usr/bin/env python3
"""Audit lexical sites in the pinned stock TypeScript compiler; never infer hotness."""
import argparse
import hashlib
import json
from pathlib import Path
import re

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("source", type=Path, help="TypeScript 050880ce checkout or archive root")
parser.add_argument("--check", action="store_true", help="compare with committed census")
args = parser.parse_args()
root = args.source / "src/compiler"
queries = {
    "tokenToString call-shaped occurrences": (None, r"\btokenToString\("),
    "getStringLiteralType call-shaped occurrences": (None, r"\bgetStringLiteralType\("),
    "checker Map construction": ("checker.ts", r"\bnew Map(?:<|\()"),
    "checker .get occurrences": ("checker.ts", r"\.get\("),
    "checker .set occurrences": ("checker.ts", r"\.set\("),
    "scanner slice calls": ("scanner.ts", r"\.slice\("),
    "factory createBaseNode call-shaped occurrences": ("factory/nodeFactory.ts", r"\bcreateBaseNode(?:<[^>]+>)?\("),
    "checker getFreshTypeOfLiteralType calls": ("checker.ts", r"\bgetFreshTypeOfLiteralType\("),
    "checker parent property occurrences": ("checker.ts", r"\.parent\b"),
}
files = sorted(root.rglob("*.ts"))
if len(files) != 77:
    raise SystemExit(f"expected 77 compiler files, got {len(files)}")
result = {
    "typescript_commit": "050880ce59e30b356b686bd3144efe24f875ebc8",
    "method": "Lexical occurrences, excluding queried function declarations and interface signatures; not dynamic frequencies or an AST-resolved receiver census. Parent occurrences include writes. A repeated line number denotes multiple matches on that line.",
    "queries": {},
}
for label, (restricted, pattern) in queries.items():
    entries = []
    for path in ([root / restricted] if restricted else files):
        lines = []
        for number, text in enumerate(path.read_text().splitlines(), 1):
            if re.search(r"\bfunction\s+(?:tokenToString|getStringLiteralType|createBaseNode|getFreshTypeOfLiteralType)\b", text):
                continue
            if re.match(r"\s*getStringLiteralType\(value: string\):", text):
                continue
            lines.extend([number] * len(re.findall(pattern, text)))
        if lines:
            entries.append({
                "file": "src/compiler/" + str(path.relative_to(root)),
                "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
                "sites": len(lines), "lines": lines,
            })
    result["queries"][label] = {"pattern": pattern, "total": sum(e["sites"] for e in entries), "files": entries}
output = Path(__file__).with_suffix(".json")
if args.check:
    if result != json.loads(output.read_text()):
        raise SystemExit("source census or fingerprint differs")
    print("source census and fingerprints: PASS")
else:
    text = json.dumps(result, indent=2)
    text = re.sub(r'("lines": \[)([\s\d,]+)(\])', lambda m: m[1] + " ".join(m[2].split()) + m[3], text)
    output.write_text(text + "\n")
    for name, entry in result["queries"].items():
        print(f"{name}: {entry['total']}")
