import json, subprocess, os
from pathlib import Path
root=Path(__file__).resolve().parent
mutants=[
 ('typed-negative-wrap','internal/native/runtime/typed_array.c','if (modulo < 0) { modulo += modulus; }','if (false && modulo < 0) { modulo += modulus; }',['go','test','./internal/native','-run','^TestTypedArrayRuntime$','-count=1','-timeout','10m']),
 ('regexp-argument-guard','internal/native/runtime/regexp_replace.c','if (!accepted) {','if (false && !accepted) {',['go','test','./internal/native','-run','RegexPrograms|ClosureConvention','-count=1','-timeout','10m']),
 ('readiness-call-invalidation','internal/lower/readiness.go','if readinessCalls(instruction) {','if false && readinessCalls(instruction) {',['go','test','./internal/lower','-count=1','-timeout','10m']),
 ('typed-kind-guard','internal/native/runtime/typed_array.c','if (array->kind != source->kind) {','if (false && array->kind != source->kind) {',['go','test','./internal/native','-run','^TestTypedArrayRuntime$','-count=1','-timeout','10m']),
]
results=[]
for name,file,before,after,cmd in mutants:
 p=Path(file);original=p.read_text();assert before in original
 try:
  p.write_text(original.replace(before,after))
  with (root/(name+'.log')).open('w') as log:
   result=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=600)
  row={'mutant':name,'file':file,'before':before,'after':after,'command':cmd,'exit':result.returncode}
  if name=='regexp-argument-guard':
   cmd=['go','test','./internal/oracle','-run','TestNativeAgreesWithNode/internal/oracle/testdata/regexp_replace/argument_guard','-count=1','-timeout','10m']
   with (root/(name+'-oracle.log')).open('w') as log: result=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=600,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'))
   row['oracle']={'command':cmd,'exit':result.returncode}
  results.append(row);print(row,flush=True)
 finally:p.write_text(original)
(root/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')
