#!/usr/bin/env python3
"""Run real lowering mutations sequentially and restore the source after each."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
source = root / "internal/lower/records_named.go"
logs = Path(os.environ.get("NAMED_INDEX_MUTANT_LOGS", "/tmp/named-index-mutants"))
logs.mkdir(parents=True, exist_ok=True)
original = source.read_text()
mutants = [
    ("index-read-type", "recordReadType", "return read, nil", "TestNamedRecordReadTypes", "named read used index type"),
    ("unrestricted-write", "recordNamedWrite", "return nil", "TestNamedRecordRefusals", "want NotYet named property named"),
    ("erase-named-contract", "sameRecordNamedContracts", "return true", "TestNamedRecordRefusals", "want NotYet seen as"),
]
for name, function, replacement, test, witness in mutants:
    at = original.index("func (l *lowering) " + function + "(")
    insert = original.index("\n", at) + 1
    try:
        source.write_text(original[:insert] + "\t" + replacement + " // deliberate mutant\n" + original[insert:])
        log = logs / (name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "./internal/lower", "-run", "^" + test + "$", "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        assert result.returncode != 0 and witness in observed, (name, result.returncode, observed)
        print(name + ": caught by " + test + " (exit " + str(result.returncode) + ")")
    finally:
        source.write_text(original)
assert source.read_text() == original
