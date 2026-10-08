#!/usr/bin/env python3
"""Executable omissions must lose both original scalar member refusals."""
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
logs=Path(__file__).resolve().parent/'logs/scalar-witnesses'
for name, filename, before, after, selection in [
 ('native-shape','internal/native/runtime/view_callables_contract.h','    if (optional && value == NULL) { return NULL; }','    return value;\n    if (optional && value == NULL) { return NULL; }','wrong-result'),
 ('javascript-shape','internal/javascript/view_callables_contract.go','    if (optional && value === undefined) return undefined;','    return value;\n    if (optional && value === undefined) return undefined;','wrong-result'),
 ('native-arity','internal/native/runtime/view_callables_contract.h','        if (recorded->arity != expected->arity) {','        if (recorded->arity != expected->arity) return value;\n        if (recorded->arity != expected->arity) {','wrong-arity'),
 ('javascript-arity','internal/javascript/view_callables_contract.go','            if (recorded.parameters.length !== expected.parameters.length) found','            if (recorded.parameters.length !== expected.parameters.length) return value;\n            if (recorded.parameters.length !== expected.parameters.length) found','wrong-arity'),
]:
 path=root/filename; original=path.read_text(); assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewCallableScalarWitnesses$/(scanner|performance)/'+selection+'$','-count=1','-timeout','5m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'error:' not in evidence,evidence
  assert 'exitCode:0' in evidence,evidence
  assert 'scanner/'+selection in evidence and 'performance/'+selection in evidence,evidence
  print(name+': both original member tests lost their named refusal',flush=True)
 finally: path.write_text(original)
