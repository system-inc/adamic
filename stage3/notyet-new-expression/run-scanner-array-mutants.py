#!/usr/bin/env python3
"""Scanner constructor dependency checks must detect unsupported acceptance."""
import json
from pathlib import Path
import subprocess
import tempfile
root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/new_dense_array.go"
original = source.read_text()
mutants = [
    ("untyped", "element, err := l.elementType(node)", "element, err := ir.Number, error(nil)", "scanner_array_notyet", "want an array of any NotYet", "an Array constructor creating holes"),
    ("holes", "argument.Type() == ir.Number || argument.Type() == ir.MaybeNumber", "false", "scanner_typed_array_notyet", "want an Array constructor creating holes NotYet, got <nil>", "got <nil>"),
]
for name, anchor, replacement, test, failure, next_stop in mutants:
    assert original.count(anchor) == 1
    with tempfile.TemporaryDirectory(prefix="scanner-array-mutant-") as scratch:
        mutated = Path(scratch) / "new_dense_array.go"
        mutated.write_text(original.replace(anchor, replacement))
        overlay = Path(scratch) / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(source): str(mutated)}}))
        log = Path("/tmp/new-expression-scanner-mutant-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/oracle", "-run", "TestNewExpressionScannerArrayDependencies/" + test, "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode != 0
        assert failure in log.read_text() and next_stop in log.read_text(), log.read_text()
        assert "clang failed" not in log.read_text()
        print(name + ": killed by exact NotYet assertion; " + next_stop, flush=True)
assert source.read_text() == original
