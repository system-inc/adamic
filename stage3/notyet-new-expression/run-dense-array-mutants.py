#!/usr/bin/env python3
"""Mutants use scratch overlays and require semantic fixture failures."""
import json
from pathlib import Path
import subprocess
import tempfile
root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/new_dense_array.go"
original = source.read_text()
mutants = [
    ("order", "literal.Elements = append(literal.Elements, value)", "literal.Elements = append([]ir.Expression{value}, literal.Elements...)", "TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_array_dense", "stdout differs"),
    ("singleton", "Elements: []ir.Expression{argument}", "Elements: []ir.Expression{}", "TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_array_dense", "stdout differs"),
    ("holes", "argument.Type() == ir.Number || argument.Type() == ir.MaybeNumber", "false", "TestNewExpressionArrayHolesRemainNotYet", "want array-hole NotYet, got <nil>"),
    ("optional-holes", "argument.Type() == ir.Number || argument.Type() == ir.MaybeNumber", "argument.Type() == ir.Number", "TestNewExpressionOptionalArrayLengthRemainsNotYet", "want array-hole NotYet, got <nil>"),
]
for name, anchor, replacement, test, failure in mutants:
    assert original.count(anchor) == 1
    with tempfile.TemporaryDirectory(prefix="dense-array-mutant-") as scratch:
        mutated = Path(scratch) / "new_dense_array.go"
        mutated.write_text(original.replace(anchor, replacement))
        overlay = Path(scratch) / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(source): str(mutated)}}))
        log = Path("/tmp/new-expression-topic-array-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/oracle", "-run", test, "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode != 0
        assert failure in log.read_text(), log.read_text()
        assert "clang failed" not in log.read_text()
        print(name + ": killed by " + failure, flush=True)
assert source.read_text() == original
