"""Run safe scalar mutants; a compiler or sanitizer error never counts as a catch."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
logs = Path(os.environ.get('ADAMIC_INTERFACE_MUTANT_LOGS', '/workspace/interface-downcast-primitive-mutants')).resolve()
assert logs != root and root not in logs.parents, 'run output must stay outside the checkout'
logs.mkdir(parents=True, exist_ok=True)
mutants = [
 ('drop-view-read', 'internal/native/emit_expressions.go', '\t\tif expression.View != "" {\n\t\t\treturn e.viewField(expression)\n\t\t}\n', '', 'TestRequiredViewFieldPrimitive/wrong_type', 'exit codes differ'),
 ('drop-view-initialization', 'internal/native/runtime/object.c', 'slot == NULL || !adamic_object_initialized(object)[cache->index]', 'slot == NULL', 'TestRequiredViewFieldPrimitive/uninitialized', 'exit codes differ'),
 ('erase-view-without-proof', 'internal/lower/readiness.go', 'if !program.CheckedFields[expression.Name] || expression.Method {', 'if true {', 'TestDefaultTaggedSourceViews/default-wrong-boolean', 'exit codes differ'),
 ('drop-view-type', 'internal/native/runtime/object.c', 'actual == wanted && wanted >= 1', 'true && wanted >= 1', 'TestRequiredViewFieldPrimitive/wrong_type', 'exit codes differ'),
 ('view-operand-twice', 'internal/native/view_fields.go', '\tobject := e.value(property.Object)\n', '\tobject := e.value(property.Object)\n\t_ = e.value(property.Object)\n', 'TestRequiredViewFieldOperandOnce', 'operand evaluated more than once: stdout differs'),
]
for name, relative, before, after, test, expected in mutants:
 source = root / relative
 original = source.read_bytes()
 try:
  text = original.decode()
  assert text.count(before) == 1, (name, 'mutation site moved')
  source.write_text(text.replace(before, after))
  with (logs / (name + '.log')).open('w') as log:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', test, '-count=1', '-v', '-timeout', '30m'], cwd=root, stdout=log, stderr=subprocess.STDOUT, env={**os.environ, 'ADAMIC_GATE_UNCACHED':'1'})
  evidence = (logs / (name + '.log')).read_text()
  assert result.returncode != 0 and expected in evidence and '--- FAIL:' in evidence, (name, evidence)
  assert not any(marker in evidence for marker in ['ERROR: AddressSanitizer', 'runtime error:', 'clang failed', 'build failed']), (name, evidence)
  print(name + ': caught by independent semantic assertion with valid C', flush=True)
 finally:
  source.write_bytes(original)
