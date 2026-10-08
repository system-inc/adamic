import json, subprocess, os
from pathlib import Path
root=Path(__file__).resolve().parent
p=Path('internal/native/runtime/typed_array.c');original=p.read_text()
probes=[('new-probes-negative-wrap','if (modulo < 0) { modulo += modulus; }','if (false && modulo < 0) { modulo += modulus; }','uint8array'),('new-probes-extra-retain','view->owner = adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array);','view->owner = adamic_retain(adamic_retain(array->owner != NULL ? array->owner : (adamic_typed_array *)array));','float64array')]
results=[]
for name,before,after,source in probes:
 assert before in original
 try:
  p.write_text(original.replace(before,after))
  with (root/(name+'-build.log')).open('w') as log:
   built=subprocess.run(['go','build','-o','/tmp/coverage-mutant-adamic','./cmd/adamic'],stdout=log,stderr=subprocess.STDOUT,timeout=180)
  assert built.returncode==0,'compiler build failure is not a killed mutant'
  env=dict(os.environ,ADAMIC_COVERAGE_COMPILER='/tmp/coverage-mutant-adamic',ADAMIC_COVERAGE_RESULTS=str(root/(name+'.json')))
  cmd=['python3',str(root/'run.py'),source]
  with (root/(name+'.log')).open('w') as log: result=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=300,env=env)
  results.append({'mutant':name,'before':before,'after':after,'program':source+'.a','command':cmd,'build_exit':built.returncode,'exit':result.returncode})
 finally:p.write_text(original)
(root/'probe-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
