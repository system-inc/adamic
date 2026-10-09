#!/usr/bin/env python3
"""Check lexical cache storage, lazy initialization and capture refusal."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/new_class_value.go"
original = source.read_text()
mutants = [
    ("eager", "return ir.Coalesce{Value: left, Fallback: right, Of: ir.Object}, nil", "_ = left; return right, nil", "TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_class_cache_local", "stdout differs"),
    ("no-store", "b.body = statements", "b.body = statements[:0]", "TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_class_cache_local", "exit codes differ"),
    ("extra-capture", "if !closed {", "if false && !closed {", "TestNewExpressionCacheCaptureRemainsNotYet", "want additional-capture NotYet, got <nil>"),
]
for name, anchor, replacement, test, failure in mutants:
    assert original.count(anchor) == 1
    with tempfile.TemporaryDirectory(prefix="constructor-local-mutant-") as scratch:
        mutated = Path(scratch) / "new_class_value.go"
        mutated.write_text(original.replace(anchor, replacement))
        overlay = Path(scratch) / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(source): str(mutated)}}))
        log = Path("/tmp/new-expression-cache-local-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/oracle", "-run", test, "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode != 0
        assert failure in log.read_text(), log.read_text()
        assert "clang failed" not in log.read_text()
        print(name + ": killed by " + failure, flush=True)
assert source.read_text() == original
