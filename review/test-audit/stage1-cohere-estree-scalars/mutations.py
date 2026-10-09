import json,subprocess,time,os
from pathlib import Path
p=Path(__file__).resolve().parent
assert len(json.loads((p/'timings.json').read_text()))==42
assert all(x['exit']==0 for x in json.loads((p/'timings.json').read_text()))
plan=json.loads((p/'frozen-mutants.json').read_text())
plan.append({'id':'P01','file':'stage1/cohere/estree/pipeline.ts','before':'export function answer(path: string, text: string): string {','after':"export function answer(path: string, text: string): string {\n    return '';"})
statuses=[]
for item in plan:
 file=Path(item['file']);original=file.read_text()
 try:
  file.write_text(original.replace(item['before'],item['after'],1))
  diff=subprocess.check_output(['git','diff','--',str(file)],text=True);(p/(item['id']+'.diff')).write_text(diff)
  env=os.environ.copy();env['ADAMIC_ESTREE_LIBRARY']='/tmp/u088/library';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u088/cache/'+item['id']
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^TestBoundedPortParser(_[0-9]{3}|Union)$']
  start=time.monotonic()
  with (p/(item['id']+'.log')).open('w') as out:result=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  statuses.append({'id':item['id'],'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'wall':time.monotonic()-start,'exit':result.returncode})
  if item['id']=='M03':
   for label,body in [('before',original),('after',file.read_text())]:
    file.write_text(body)
    input=Path('/tmp/u088/recovery-witness.ts');input.write_text('type X = { ) };')
    with (p/('M03-witness-'+label+'.log')).open('w') as out:subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','stage1/cohere/estree/main.ts',str(input)],stdout=out,stderr=subprocess.STDOUT,timeout=10)
  (p/'mutation-status.json').write_text(json.dumps(statuses,indent=2))
  print(item['id'],statuses[-1],flush=True)
 finally:file.write_text(original)
