#!/usr/bin/env python3
"""Prove that eager initialization cannot pass the constructor || fixture."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/new_class_value.go"
text = source.read_text()
anchor = "return ir.Coalesce{Value: left, Fallback: right, Of: ir.Object}, nil"
assert text.count(anchor) == 1
text = text.replace(anchor, "if binary.OperatorToken.Kind == ast.KindBarBarToken { return right, nil }; " + anchor)
with tempfile.TemporaryDirectory(prefix="constructor-or-mutant-") as scratch:
    mutated = Path(scratch) / "new_class_value.go"
    mutated.write_text(text)
    overlay = Path(scratch) / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(source): str(mutated)}}))
    log = Path("/tmp/new-expression-cache-or-mutant.log")
    with log.open("w") as output:
        result = subprocess.run(["go", "test", "-overlay=" + str(overlay),
            "./internal/oracle", "-run", "TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_class_cache_or",
            "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode != 0
    assert "stdout differs" in log.read_text()
    assert "clang failed" not in log.read_text()
    print("eager || cache initializer killed by source Node stdout comparison")
