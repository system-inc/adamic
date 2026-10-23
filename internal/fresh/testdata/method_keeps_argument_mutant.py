"""Prove TestMethodKeepsArgument detects arguments incorrectly left confined.

Run with the cloud toolchain environment sourced. Output stays outside the tree.
The production source is restored even if the bounded test fails or times out.
"""
from pathlib import Path
import os
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
source = root / "internal/fresh/fresh.go"
original = source.read_text()
old = "\tfor _, operand := range operands {\n\t\ta.state.escape(operand)\n\t}"
assert original.count(old) == 1, "call escape seam changed"
log = Path(tempfile.mkdtemp(prefix="method-keeps-argument-mutant-")) / "test.log"
try:
    source.write_text(original.replace(old, "\t// Mutant: arguments incorrectly remain confined."))
    with log.open("w") as output:
        result = subprocess.run(
            ["go", "test", "-p", "1", "./internal/fresh", "-run",
             "^TestMethodKeepsArgument$", "-count=1", "-v", "-timeout", "55s"],
            cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=58,
            env=dict(os.environ, GOMAXPROCS="4"),
        )
finally:
    source.write_text(original)
text = log.read_text()
assert result.returncode == 1, f"mutant survived or failed to run: {log}"
assert "method-kept argument must escape before the push" in text, text
assert "Proven:true" in text, text
print(f"TestMethodKeepsArgument caught confined-argument mutant: {log}")
