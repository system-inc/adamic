#!/usr/bin/env python3
"""Prove that accepting Uint16 through Int32 cannot pass the refusal fixture."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/typed_arrays.go"
text = source.read_text()
old = 'case ir.Int32Array:\n\t\treturn "Int32Array"'
assert text.count(old) == 1
text = text.replace(old, 'case ir.Int32Array:\n\t\treturn "Uint16Array"')
start = text.index("if node.Kind == ast.KindNewExpression")
old = '"Int16Array", "Uint16Array", "Uint32Array"'
assert text[start:].count(old) == 1
text = text[:start] + text[start:].replace(old, '"Int16Array", "Uint32Array"', 1)
with tempfile.TemporaryDirectory(prefix="uint16-notyet-mutant-") as scratch:
    mutated = Path(scratch) / "typed_arrays.go"
    mutated.write_text(text)
    overlay = Path(scratch) / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(source): str(mutated)}}))
    log = Path("/tmp/new-expression-drop-mutant.log")
    with log.open("w") as output:
        result = subprocess.run(["go", "test", "-overlay=" + str(overlay),
            "./internal/oracle", "-run", "^TestNewExpressionUint16RemainsNotYet$",
            "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode != 0
    assert "want Uint16Array element-type NotYet, got <nil>" in log.read_text()
    print("Int32 acceptance mutant killed by exact Uint16Array NotYet assertion")
