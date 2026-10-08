#!/usr/bin/env python3
"""Isolated, compiling mutants for the stopped join boundary and query."""
import json
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[4]
out = pathlib.Path(os.environ.get("ADAMIC_STEP37_EVIDENCE", "/workspace/scratch/step37/mutants"))
out.mkdir(parents=True, exist_ok=True)
cases = [
    ("join-guard", "internal/lower/graph_types.go", "return f.parallelJoinResults(modules)", "return nil", "./internal/oracle", "^TestStep37JoinRefusesUnprovenGraph$/private_graph$", "graph handoff reached emission: <nil>"),
    ("worker-owner", "internal/fresh/ownership.go", "case root.Name != owner:", "case false: // mutant ignores surviving worker roots", "./internal/fresh", "^TestStep37JoinInventory$", "worker reference or incomplete inventory crossed join"),
    ("incomplete-inventory", "internal/fresh/ownership.go", "if !s.Complete {", "if false { // mutant ignores incomplete inventory", "./internal/fresh", "^TestStep37JoinInventory$", "worker reference or incomplete inventory crossed join"),
]
for name, relative, before, after, package, selection, catcher in cases:
    original = root / relative
    source = original.read_text()
    assert source.count(before) == 1, (name, "lost unique anchor")
    changed = out / (name + ".go")
    changed.write_text(source.replace(before, after, 1))
    overlay = out / (name + ".json")
    overlay.write_text(json.dumps({"Replace": {str(original): str(changed)}}))
    log = out / (name + ".log")
    with log.open("w") as stream:
        result = subprocess.run(["go", "test", "-overlay", str(overlay), package, "-run", selection, "-count=1", "-v", "-timeout", "10m"], cwd=root, stdout=stream, stderr=subprocess.STDOUT, timeout=660)
    text = log.read_text()
    assert result.returncode != 0 and catcher in text and "build failed" not in text and "undefined:" not in text, (name, result.returncode, text)
    print(name + ": caught by " + catcher + "; log=" + str(log))
print("3 compiling mutants caught; production files were never changed")
