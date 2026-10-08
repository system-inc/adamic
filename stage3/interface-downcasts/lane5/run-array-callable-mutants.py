#!/usr/bin/env python3
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
logs=Path(__file__).resolve().parent/'logs/aggregate'
mutants=[
 ('array-native-arity','internal/native/runtime/view_callables_contract.h','        if (recorded->arity != expected->arity) {','        if (recorded->arity != expected->arity) return value;\n        if (recorded->arity != expected->arity) {','^TestCheckedViewCallableArrays$/.*/^wrong-arity$'),
 ('array-javascript-arity','internal/javascript/view_callables_contract.go','            if (recorded.parameters.length !== expected.parameters.length) found','            if (recorded.parameters.length !== expected.parameters.length) return value;\n            if (recorded.parameters.length !== expected.parameters.length) found','^TestCheckedViewCallableArrays$/.*/^wrong-arity$'),
 ('array-payload-descriptors','internal/lower/view_callables_aggregate.go','id, err := l.viewContract(payload.node, payload.proven)','if l.viewArrayElementType(payload.proven) != nil { continue }\n id, err := l.viewContract(payload.node, payload.proven)','^TestCheckedViewCallableArrays$/.*/^wrong-element$'),
]
for name,filename,before,after,selection in mutants:
 path=root/filename;original=path.read_text();assert original.count(before)==(2 if name=='array-payload-descriptors' else 1)
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run',selection,'-count=1','-timeout','5m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'clang failed:' not in evidence,evidence
  assert 'exitCode:0' in evidence,evidence
  for group in ['call-expression/','inline-expressions/']:assert group in evidence,evidence
  print(name+': executable mutant killed for both original member fixtures',flush=True)
 finally:path.write_text(original)
