"""Run contextual and lazy-operand mutants; restore each source file on exit."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[3]
source = root / "internal/lower/generic_function_value.go"
mutants = [
 ("direct-only", "contextual := l.concrete(l.checker.GetContextualType(node, checker.ContextFlagsNone))", "if !viewSite(node) { return nil }\n\tcontextual := l.concrete(l.checker.GetContextualType(node, checker.ContextFlagsNone))", "./internal/lower", "TestGenericFunctionValueOperandContexts", ["missing concrete operand signature"]),
 ("or-swapped", "ir.Coalesce{Value: left, Fallback: fit(right, ir.Closure), Of: ir.Closure}", "ir.Coalesce{Value: right, Fallback: left, Of: ir.Closure}", "./internal/oracle", "TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_value_arms", ["stdout differs", "JavaScript backend: stdout differs"]),
 ("and-inverted", "Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: left}}", "Condition: ir.IsUndefined{Value: left}", "./internal/oracle", "TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_value_arms", ["stdout differs", "JavaScript backend: stdout differs"]),
]
original = source.read_text()
for name, old, new, package, test, catchers in mutants:
 assert original.count(old) == 1, (name, original.count(old))
 try:
  source.write_text(original.replace(old, new))
  log = Path("/tmp") / ("generic-arms-mutant-" + name + ".log")
  with log.open("w") as output:
   result = subprocess.run(["go", "test", package, "-run", "^" + test, "-count=1", "-timeout", "30m"], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"), stdout=output, stderr=subprocess.STDOUT)
  text = log.read_text()
  assert result.returncode != 0 and all(catcher in text for catcher in catchers), (name, text)
  assert "clang:" not in text and "runtime error:" not in text, (name, text)
  print(name + ": caught by " + ", ".join(catchers))
 finally:
  source.write_text(original)
