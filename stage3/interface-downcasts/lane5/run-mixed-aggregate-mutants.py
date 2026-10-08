#!/usr/bin/env python3
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
logs=Path(__file__).resolve().parent/'logs/aggregate'
selection='^TestCheckedViewCallable(ElementAccess|PropertyWitnesses)$'
mutants=[
 ('mixed-native-arity','internal/native/runtime/view_callables_contract.h','        if (recorded->arity != expected->arity) {','        if (recorded->arity != expected->arity) return value;\n        if (recorded->arity != expected->arity) {',selection+'/(property-access|property-assignment|wrong-arity)/^wrong-arity$'),
 ('mixed-javascript-arity','internal/javascript/view_callables_contract.go','            if (recorded.parameters.length !== expected.parameters.length) found','            if (recorded.parameters.length !== expected.parameters.length) return value;\n            if (recorded.parameters.length !== expected.parameters.length) found',selection+'/(property-access|property-assignment|wrong-arity)/^wrong-arity$'),
 ('mixed-native-parameters','internal/native/runtime/view_callables_contract.h','            found = "function with incompatible parameter representations";','            return value;',selection+'/(property-access|property-assignment|wrong-members)/^wrong-members$'),
 ('mixed-javascript-parameters','internal/javascript/view_callables_contract.go','            else found = "function with incompatible parameter representations";','            else return value;',selection+'/(property-access|property-assignment|wrong-members)/^wrong-members$'),
 ('mixed-boxed-field','internal/lower/view_callables_boxing.go','return known && of == ir.Union && l.viewCallableAggregateType(declared) && l.viewCallableBoxedRepresentation(declared)','_ = of; _ = known; _ = declared\n return false','^TestCheckedViewCallableElementAccess$/^good$'),
]
for name,filename,before,after,tests in mutants:
 path=root/filename;original=path.read_text();assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run',tests,'-count=1','-timeout','5m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'clang failed:' not in evidence,evidence
  if name=='mixed-boxed-field': assert 'exit codes differ' in evidence and 'exitCode:-1' in evidence,evidence
  else:
   assert 'exitCode:0' in evidence,evidence
   for group in ['property-access','property-assignment','TestCheckedViewCallableElementAccess/']: assert group in evidence,evidence
  print(name+': executable mutant killed',flush=True)
 finally:path.write_text(original)
