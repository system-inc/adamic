#!/usr/bin/env python3
"""Prove source dispatch checks fail semantically; always restore each file."""
from pathlib import Path
import subprocess
root = Path(__file__).resolve().parents[3]
logs = Path(__file__).resolve().parent / 'logs/source-mutants'
logs.mkdir(parents=True, exist_ok=True)
mutants = [
 ('lower-skip-read-contract', 'internal/lower/object.go', 'l.prepareViewCallableProperty(node, declared, &property)', '// mutant: omit read contract', 'TestCheckedViewCallableSourceContracts/number-wrong-arity'),
 ('native-skip-shape-dispatch', 'internal/native/view_fields.go', 'value = e.emitViewCallableCertificate(property, callable)', 'value = callable', 'TestCheckedViewCallableSourceContracts/number-wrong-arity'),
 ('native-accept-wrong-arity', 'internal/native/runtime/view_callables_contract.h', 'recorded->arity != expected->arity', 'false', 'TestCheckedViewCallableSourceContracts/number-wrong-arity'),
 ('javascript-accept-wrong-arity', 'internal/javascript/view_callables_contract.go', 'recorded.parameters.length !== expected.parameters.length', 'false', 'TestCheckedViewCallableSourceContracts/number-wrong-arity'),
 ('javascript-skip-shape', 'internal/javascript/view_callables_contract.go', '    if (optional && value === undefined)', '    if (value !== undefined) return value;\n    if (optional && value === undefined)', 'TestCheckedViewCallableSourceContracts/number-wrong-arity'),
 ('native-skip-method-signature', 'internal/native/view_callables_methods.go', '(void)adamic_view_callable_shape(&%s.heap, %s, %s, %s, false);', '(void)&%s.heap; (void)%s; (void)%s; (void)%s;', 'TestCheckedViewCallableMethods/method-wrong-arity'),
 ('lower-drop-marker-arity-proof', 'internal/lower/view_callables_marker.go', ' || len(from[0].Parameters()) != 0', '', 'TestCheckedViewCallableMarker/discarded-required'),
 ('lower-allow-write-back', 'internal/lower/invariance.go', 'if !fresh && !l.checker.IsReadonlySymbol(viewed) && (!l.enumAssignable(target, source) || !l.checker.IsTypeAssignableTo(target, source)) {', 'if false {', 'TestCheckedViewCallableMarker/write-back'),
]
for name, filename, before, after, test in mutants:
 path = root / filename
 original = path.read_text()
 assert original.count(before) == 1, name
 try:
  path.write_text(original.replace(before, after))
  with (logs / (name + '.log')).open('w') as output:
   result = subprocess.run(['go', 'test', './internal/oracle', '-run', '^' + test + '$', '-count=1', '-timeout', '5m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  evidence = (logs / (name + '.log')).read_text()
  assert result.returncode != 0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'error:' not in evidence, name + ': not semantically killed: ' + evidence
  print(name + ': caught by ' + test, flush=True)
 finally:
  path.write_text(original)
