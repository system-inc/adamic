#!/usr/bin/env python3
"""Semantic mutants of reached-read and immutable producer certificate adapters."""
from pathlib import Path
import subprocess
root = Path(__file__).resolve().parents[3]
logs = Path(__file__).resolve().parent / 'logs/read-adapter-mutants'
logs.mkdir(parents=True, exist_ok=True)
mutants = [
 ('native-forge-producer', 'internal/native/view_callables_signature.go',
  'emitViewCallableShape(value, recorded, expected,',
  'emitViewCallableShape(value, "(" + recorded + " != NULL ? " + expected + " : NULL)", expected,',
  './internal/native', 'TestViewCallableProducerCertificateNative/wrong-arity'),
 ('javascript-forge-producer', 'internal/javascript/view_callables_signature.go',
  'emitViewCallableShape("value", recorded, expected,',
  'emitViewCallableShape("value", "(" + recorded + ", " + expected + ")", expected,',
  './internal/javascript', 'TestViewCallableProducerCertificateNode/wrong-arity'),
 ('lower-eager-signature-descendants', 'internal/lower/view_callables_read.go',
  'of, known := l.representation(child)',
  'return l.viewContract(node, child)\n        of, known := l.representation(child)',
  './internal/lower', 'TestPrepareViewCallableRead/interface_Result'),
]
for name, filename, old, new, package, test in mutants:
 path = root / filename
 original = path.read_text()
 assert original.count(old) == 1, name
 try:
  path.write_text(original.replace(old, new))
  with (logs / (name + '.log')).open('w') as output:
   result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-timeout', '5m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
  evidence = (logs / (name + '.log')).read_text()
  assert result.returncode != 0 and '--- FAIL:' in evidence and '[build failed]' not in evidence, name + ' not semantically killed: ' + evidence
  print(name + ': caught by ' + test)
 finally:
  path.write_text(original)
