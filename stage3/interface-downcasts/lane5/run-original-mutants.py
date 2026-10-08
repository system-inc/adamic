#!/usr/bin/env python3
"""Both original witness pairs must catch an arity erasure in each backend."""
from pathlib import Path
import subprocess
root = Path(__file__).resolve().parents[3]
logs = Path(__file__).resolve().parent / 'logs/original-mutants'
logs.mkdir(parents=True, exist_ok=True)
for name, filename, before in [
 ('native', 'internal/native/runtime/view_callables_contract.h', 'recorded->arity != expected->arity'),
 ('javascript', 'internal/javascript/view_callables_contract.go', 'recorded.parameters.length !== expected.parameters.length'),
]:
 path = root / filename
 original = path.read_text()
 assert original.count(before) == 1
 try:
  path.write_text(original.replace(before, 'false'))
  with (logs / (name+'.log')).open('w') as output:
   result = subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewCallable(Canonical|Watcher)Witness$/wrong-arity$','-count=1','-timeout','5m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  evidence = (logs/(name+'.log')).read_text()
  assert result.returncode != 0 and '[build failed]' not in evidence and 'error:' not in evidence, evidence
  for pair in ['Canonical', 'Watcher']:
   assert '--- FAIL: TestCheckedViewCallable'+pair+'Witness/wrong-arity' in evidence, evidence
  print(name+': arity erasure caught by both original witness pairs',flush=True)
 finally:
  path.write_text(original)
