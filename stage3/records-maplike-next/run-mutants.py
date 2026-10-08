#!/usr/bin/env python3
"""Verify the refusals where records meet area-next and restore each mutation."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/expression.go"
logs = Path(os.environ.get("RECORDS_NEXT_MUTANT_LOGS", "/tmp/maplike-next-merge-mutants"))
logs.mkdir(parents=True, exist_ok=True)
original = source.read_text()
mutants = [
    ("logical-record-conversion", original.replace(
        '\t\t\tif left.Type() == ir.Record && of != ir.Record && of != ir.Union && !of.IsMaybe() {',
        '\t\t\tif false {', 1), 'want "logical record operand"'),
    ("nullable-record-container", original.replace(
        ' || of == ir.Record || of == ir.String', ' || of == ir.String', 1).replace(
        '\t\tif shared == ir.Record && l.includesNull(proven) {',
        '\t\tif false {', 1), 'want "value of type"'),
]
for name, changed, witness in mutants:
    assert changed != original, name
    try:
        source.write_text(changed)
        with (logs / (name + ".log")).open("w") as output:
            result = subprocess.run([
                "go", "test", "./internal/lower", "-run", "^TestRecordRefusals$",
                "-count=1", "-timeout", "10m",
            ], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observed = (logs / (name + ".log")).read_text()
        assert result.returncode != 0 and witness in observed, (name, result.returncode, observed)
        print(name + ": caught by TestRecordRefusals (exit " + str(result.returncode) + ")")
    finally:
        source.write_text(original)
assert source.read_text() == original
