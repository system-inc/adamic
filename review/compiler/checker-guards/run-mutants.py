#!/usr/bin/env python3
"""Apply each diff, check its named test, restore, and check that test again."""
import difflib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[3]
OUT = Path(__file__).resolve().parent

CASES = [
    ("M1", "TestGuardSymbolOriginFFFF", "invalid frame length"),
    ("M4", "TestGuardStrictNullChecks", "strict-null-checks metadata"),
    ("M5", "TestGuardSingleTypeRoot", "type root count"),
]
EXTRA = [
    ("symbol-name", "metadata.go", "out.text(symbol.Name)", 'out.text("mutant")', "TestGuardWriteSymbolOrigin", "symbol declaration origin"),
    ("property-annotation", "metadata.go", "out.text(annotation)", 'out.text("mutant" + annotation)', "TestGuardWritePropertyInfo", "method parameter metadata"),
    ("diagnostic-separator", "program.go", 'return strings.Join(messages, "; ")', 'return strings.Join(messages, " | ")', "TestGuardDiagnosticText", "diagnostic text"),
    ("query-type", "program.go", "Type: typeChecker.TypeToString(typeChecker.GetTypeAtLocation(node))", 'Type: "mutant"', "TestGuardQuery", "binary expression query"),
    ("parts-name", "program.go", "name := typeChecker.TypeToString(part)", 'name := "mutant"', "TestGuardTypeParts", "number type parts"),
    ("scope-order", "scopes.go", "sort.Strings(names)", "sort.Sort(sort.Reverse(sort.StringSlice(names)))", "TestGuardScopeTables", "binder scope table"),
]

for name, filename, before, after, test, catcher in EXTRA:
    relative = "bridge/tsgo/checker/" + filename
    original = (ROOT / relative).read_text()
    assert original.count(before) == 1, (name, "mutation site moved")
    changed = original.replace(before, after, 1)
    diff = "".join(difflib.unified_diff(original.splitlines(True), changed.splitlines(True), fromfile="a/" + relative, tofile="b/" + relative))
    (OUT / (name + ".diff")).write_text(diff)
    CASES.append((name, test, catcher))


def run(test, destination):
    command = ["go", "test", "./bridge/tsgo/checker", "-run", "^" + test + "$", "-count=1", "-v", "-timeout", "90s"]
    with destination.open("w") as log:
        result = subprocess.run(command, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT, timeout=150)
    return result.returncode, destination.read_text(), command


results = []
for name, test, catcher in CASES:
    patch = OUT / (name + ".diff")
    subprocess.run(["git", "apply", "--check", str(patch)], cwd=ROOT, check=True, timeout=30)
    subprocess.run(["git", "apply", str(patch)], cwd=ROOT, check=True, timeout=30)
    try:
        code, text, command = run(test, OUT / (name + "-mutant.log"))
        assert code != 0 and "--- FAIL: " + test in text and catcher in text, (name, text)
        assert "build failed" not in text and "panic:" not in text and "timed out" not in text, (name, text)
    finally:
        subprocess.run(["git", "apply", "-R", str(patch)], cwd=ROOT, check=True, timeout=30)
    restored, text, _ = run(test, OUT / (name + "-restored.log"))
    assert restored == 0 and "--- PASS: " + test in text, (name, text)
    results.append({"mutant": name, "test": test, "catcher": catcher, "mutant_exit": code, "restored_exit": restored, "command": command})
    (OUT / "mutants.json").write_text(json.dumps(results, indent=2) + "\n")
    print(name + ": named assertion fails; restored test passes", flush=True)
