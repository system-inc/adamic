#!/usr/bin/env python3
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
logs=Path(__file__).resolve().parent/'logs/aggregate'
mutants=[
 ('numeric-native-arity','internal/native/runtime/view_callables_contract.h','        if (recorded->arity != expected->arity) {','        if (recorded->arity != expected->arity) return value;\n        if (recorded->arity != expected->arity) {','wrong-arity','exitCode:0'),
 ('numeric-javascript-arity','internal/javascript/view_callables_contract.go','            if (recorded.parameters.length !== expected.parameters.length) found','            if (recorded.parameters.length !== expected.parameters.length) return value;\n            if (recorded.parameters.length !== expected.parameters.length) found','wrong-arity','exitCode:0'),
 ('numeric-missing-argument','internal/native/census_optional_values.go','missing := zero(of)','missing := zero(of)\n if of == ir.MaybeNumber { missing = "(adamic_maybe_number){true, 0.0}" }','optional-values','disagreement'),
 ('numeric-message-capacity','internal/native/runtime/object.c','2 * strlen(type)','strlen(type)','wrong-value','want "adamic: panic:'),
]
for name,filename,before,after,probe,needle in mutants:
 path=root/filename;original=path.read_text();assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewCallableNumericLiteral$/^'+probe+'$','-count=1','-timeout','5m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'clang failed:' not in evidence,evidence
  if needle=='disagreement': assert 'stdout differs' in evidence,evidence
  else:assert needle in evidence,evidence
  print(name+': executable mutant killed by '+probe,flush=True)
 finally:path.write_text(original)
