#!/usr/bin/env python3
"""Erase each independent proof and require its pinned negative control to fail."""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
evidence = Path(__file__).resolve().parent / "evidence"
mutants = {
    "callback": ("internal/lower/predicates_proof.go", "func (l *lowering) proveInferredPredicate(function *ast.Node, target *checker.Type) bool {", "\n\treturn true\n", "TestParserCallbackLie"),
}
for name in sys.argv[1:] or mutants:
    relative, signature, replacement, test = mutants[name]
    path = root / relative
    original = path.read_text()
    assert original.count(signature) == 1
    start = original.index(signature) + len(signature)
    # Insert an unconditional return. Go permits unreachable statements, so this
    # removes only the verifier, with no build-warning failure as a substitute.
    try:
        path.write_text(original[:start] + replacement + original[start:])
        with (evidence / ("mutant-" + name + ".log")).open("w") as output:
            result = subprocess.run(["go", "test", "./internal/lower", "-run", "^" + test + "$", "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        text = (evidence / ("mutant-" + name + ".log")).read_text()
        assert result.returncode != 0 and "--- FAIL: " + test in text and "[build failed]" not in text, text
        print(name + ": caught by " + test)
    finally:
        path.write_text(original)
