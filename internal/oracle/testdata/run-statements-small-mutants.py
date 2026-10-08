#!/usr/bin/env python3
"""Run independent lowering mutants, restoring every source before continuing."""
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path("/tmp/statements-small-mutants")
logs.mkdir(exist_ok=True)
mutants = [
    ("prefix-write", "internal/lower/expression.go", "l.result.Functions[index].Body = append(body, ir.Return{Value: value})", "l.result.Functions[index].Body = []ir.Statement{ir.Return{Value: value}}; _ = body", "statements_small_prefix"),
    ("prefix-result", "internal/lower/expression.go", "l.result.Functions[index].Body = append(body, ir.Return{Value: value})", "l.result.Functions[index].Body = append(body, ir.Return{Value: ir.Binary{Operator: ir.Subtract, Left: value, Right: ir.NumberConstant{Value: 1}}})", "statements_small_prefix"),
]
for name, relative, original, replacement, fixture in mutants:
    path = root / relative
    clean = path.read_text()
    assert clean.count(original) == 1, name
    try:
        path.write_text(clean.replace(original, replacement))
        with (logs / (name + ".log")).open("w") as log:
            result = subprocess.run(["go", "test", "./internal/oracle", "-run", "TestNativeAgreesWithNode/internal/oracle/testdata/" + fixture, "-count=1", "-timeout", "10m"], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"), stdout=log, stderr=subprocess.STDOUT)
        output = (logs / (name + ".log")).read_text()
        assert result.returncode != 0 and "stdout" in output and "FAIL" in output, name + ": not caught by output comparison"
        print(name + ": caught by stdout comparison")
    finally:
        path.write_text(clean)
