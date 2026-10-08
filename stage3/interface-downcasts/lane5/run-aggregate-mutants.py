#!/usr/bin/env python3
from pathlib import Path
import subprocess
root=Path(__file__).resolve().parents[3]
logs=Path(__file__).resolve().parent/'logs/aggregate'
for name,filename,before,after,selection in [
 ('native-arity','internal/native/runtime/view_callables_contract.h','        if (recorded->arity != expected->arity) {','        if (recorded->arity != expected->arity) return value;\n        if (recorded->arity != expected->arity) {','^TestCheckedViewCallableAggregateWitnesses$/.*/^wrong-arity$'),
 ('javascript-arity','internal/javascript/view_callables_contract.go','            if (recorded.parameters.length !== expected.parameters.length) found','            if (recorded.parameters.length !== expected.parameters.length) return value;\n            if (recorded.parameters.length !== expected.parameters.length) found','^TestCheckedViewCallableAggregateWitnesses$/.*/^wrong-arity$'),
 ('payload-read-registration','internal/lower/view_callables_aggregate.go','program.CheckedFields[field.Name] = true','// omit transitive read registration','^TestCheckedViewCallableAggregateWitnesses$/(expression-statement|void-zero|this)/^wrong-payload$'),
]:
 path=root/filename;original=path.read_text();assert original.count(before)==1
 try:
  path.write_text(original.replace(before,after))
  with (logs/(name+'.log')).open('w') as log:
   result=subprocess.run(['go','test','./internal/oracle','-run',selection,'-count=1','-timeout','5m'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  evidence=(logs/(name+'.log')).read_text()
  assert result.returncode!=0 and '--- FAIL:' in evidence and '[build failed]' not in evidence and 'clang failed:' not in evidence,evidence
  assert 'exitCode:0' in evidence,evidence
  for group in (['expression-statement','diagnostic-add','void-zero','this','emit-helper','node-check-flag'] if name!='payload-read-registration' else ['expression-statement','void-zero','this']):assert group+'/' in evidence,evidence
  print(name+': executable mutant killed in every selected original member fixture',flush=True)
 finally:path.write_text(original)

# The graph hook has its own control: a disjoint ordinary return stays admitted.
path=root/'internal/lower/view_callables_aggregate.go'
original=path.read_text()
before='add(graph.ReachingAllocations(call))'
assert original.count(before)==1
try:
 path.write_text(original.replace(before,'// omit callable result demand'))
 with (logs/'result-demand-omission.log').open('w') as log:
  result=subprocess.run(['go','test','./internal/lower','-run','^TestViewCallableAggregateReturnedDemand$','-count=1'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
 evidence=(logs/'result-demand-omission.log').read_text()
 assert result.returncode!=0 and 'returned allocation lost demanded refusal: <nil>' in evidence and '[build failed]' not in evidence,evidence
 print('result-demand-omission: returned allocation refusal lost; ordinary disjoint control retained',flush=True)
finally:path.write_text(original)
