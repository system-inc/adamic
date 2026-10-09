"""Run independent one-line lowering mutants, restoring sources even on failure."""
from pathlib import Path
import difflib
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = Path(__file__).resolve().parent
lower = root / "internal/lower/iteration.go"
tests = root / "internal/lower/iteration_test.go"
original_lower = lower.read_text()
converted_tests = tests.read_text()
old_tests = subprocess.check_output(["git", "show", "96e5c1dc:internal/lower/iteration_test.go"], cwd=root, text=True)
mutants = [
    ("unbound-method-call", "property.Method = true", "property.Method = false", "TestLiteralMethodViewsDoNotLoseThis"),
    ("wrong-method-result", "err := l.lowerFunction(index, node, this)", "err := l.lowerFunction(index, node, this); l.result.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 2}}}", "TestLiteralMethodSignatureViewsDoNotLoseThis"),
]

def run(name, pattern):
    command = ["go", "test", "./internal/lower", "-run", "^" + pattern + "$", "-count=1", "-timeout", "90s", "-v"]
    with (evidence / (name + ".log")).open("w") as log:
        result = subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=150)
    return result.returncode

try:
    for name, before, after, target in mutants:
        lower = root / ("internal/lower/expression.go" if name == "unbound-method-call" else "internal/lower/iteration.go")
        original_lower = lower.read_text()
        assert original_lower.count(before) == 1
        changed = original_lower.replace(before, after)
        (evidence / (name + ".diff")).write_text("".join(difflib.unified_diff(original_lower.splitlines(True), changed.splitlines(True), fromfile="a/" + str(lower.relative_to(root)), tofile="b/" + str(lower.relative_to(root)))))
        lower.write_text(changed)
        tests.write_text(old_tests)
        old_code = run(name + "-old", "TestLiteralMethodViewsDoNotLoseThis")
        tests.write_text(converted_tests)
        new_code = run(name + "-converted", target)
        output = (evidence / (name + "-converted.log")).read_text()
        print(f"{name}: old exit {old_code}, converted exit {new_code}", flush=True)
        assert old_code == 0 and new_code == 1 and "JavaScript backend stdout" in output, output
        lower.write_text(original_lower)
finally:
    lower.write_text(original_lower)
    tests.write_text(converted_tests)
