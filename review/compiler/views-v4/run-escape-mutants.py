"""Pin compiler-half guard omissions; adapter mutants remain pending."""
from pathlib import Path
import difflib
import subprocess

root = Path(__file__).resolve().parents[3]
path = root / "internal/lower/view_callable_escape.go"
original = path.read_bytes()
text = original.decode()
cases = [
    ("escape-argument", "TestV4EscapeArgumentRefusal", "if graph.viewCallableEscapeProven(read, reaches) {", "if true {"),
    ("escape-result", "TestV4EscapeResultRefusal", "if graph.viewCallableEscapeProven(read, reaches) {", "if true {"),
    ("escape-passed", "TestV4EscapePassedRefusal", "if graph.viewCallableEscapeProven(read, reaches) {", "if true {"),
    ("escape-returned", "TestV4EscapeReturnedRefusal", "if graph.viewCallableEscapeProven(read, reaches) {", "if true {"),
    ("escape-mixed", "TestV4EscapeMixedProducerRefusal", "if graph.viewCallableEscapeProven(read, reaches) {", "if true {"),
]
line = next(line for line in text.splitlines() if line.startswith("\treturn &NotYet{"))
cases.append(("escape-typescript", "TestV4EscapeTypeScriptBoundary", line, "\treturn nil"))
for name, test, before, after in cases:
    assert text.count(before) == 1, name
    changed = text.replace(before, after)
    evidence = root / "review/compiler/views-v4" / name
    evidence.with_suffix(".patch").write_text("".join(difflib.unified_diff(text.splitlines(True), changed.splitlines(True), fromfile=str(path.relative_to(root)), tofile=str(path.relative_to(root)))))
    try:
        path.write_text(changed)
        with evidence.with_suffix(".log").open("w") as output:
            result = subprocess.run(["timeout", "90", "go", "test", "./internal/oracle", "-run", "^" + test + "$", "-count=1", "-v", "-timeout", "90s"], cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=90)
        observation = evidence.with_suffix(".log").read_text()
        assert result.returncode == 1, observation
        for backend in ["native", "sanitized", "javascript"]:
            assert "guard mutant admitted in " + backend + ": exit 0" in observation, observation
        assert "with path and fix: <nil>" in observation or "explicitly unavailable: <nil>" in observation, observation
        assert "Sanitizer" not in observation, observation
        print(name + ": caught by missing compiler boundary; clean Node-equivalent native, sanitized, and JavaScript execution")
    finally:
        path.write_bytes(original)
