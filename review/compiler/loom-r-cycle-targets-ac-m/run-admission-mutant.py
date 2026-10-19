import difflib
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = Path(__file__).resolve().parent
path = root / "internal/lower/cycles.go"
original = path.read_text()
start = "\t\t\tif f.l.result.Locals[node.cell-1].Type == ir.Closure {"
end = "\t\t\tif proven := f.l.localTypes[node.cell-1]; proven != nil {"
a = original.index(start)
b = original.index(end, a)
mutated = original[:a] + original[b:]
start = "\t\t\t// Stored values and calls use the same producer graph."
end = "\t\tcase f.l.checker.IsArrayType(proven)"
a = mutated.index(start)
b = mutated.index(end, a)
old = """\t\t\tfor _, closure := range f.l.closureRecords {
\t\t\t\tif !f.l.checker.IsTypeAssignableTo(closure.proven, proven) {
\t\t\t\t\tcontinue
\t\t\t\t}
\t\t\t\tfor _, local := range f.l.result.Functions[closure.function].Environment {
\t\t\t\t\tqueue = append(queue, cycleNode{cell: local + 1})
\t\t\t\t}
\t\t\t}
"""
mutated = mutated[:a] + old + mutated[b:]
(evidence / "type-matched-captures.patch").write_text("".join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile="internal/lower/cycles.go", tofile="internal/lower/cycles.go")))
try:
    path.write_text(mutated)
    with (evidence / "type-matched-captures.log").open("w") as log:
        result = subprocess.run(["go", "test", "./internal/lower", "./internal/oracle", "-run", "^(TestNamespaceAmbientCwdInitialization|TestClosureCycleAmbientCwd)$", "-count=1", "-timeout", "90s", "-parallel", "4", "-v"], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=120)
    output = (evidence / "type-matched-captures.log").read_text()
    if result.returncode != 1 or "adamic/cycle-capable" not in output or "build failed" in output or "--- FAIL: TestNamespaceAmbientCwdInitialization" not in output or "--- FAIL: TestClosureCycleAmbientCwd" not in output:
        raise SystemExit("old-rule mutant was not caught by both admission tests")
    print("type-matched-captures: caught by both cwd admission tests")
finally:
    path.write_text(original)
