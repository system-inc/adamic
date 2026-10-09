import json,subprocess,time,os
from pathlib import Path
from selection import bounded
p=Path('review/test-audit/internal-oracle-wasi');data=[]
for mid in [f'M{i:02d}' for i in range(1,10)]+['P01']:
 for label,selection in [('bounded',bounded),('emission','^TestWASIEmission$')]:
  env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/workspace/u072-products/'+mid
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',selection]
  start=time.monotonic();log=p/f'{mid}-{label}.log'
  with open(log,'w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in open(log):
   try:events.append(json.loads(l))
   except:pass
  failed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')))
  ran=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='run' and e.get('Test')))
  end=[e for e in events if e.get('Action') in ['pass','fail'] and 'Test' not in e]
  item=dict(id=mid,part=label,command='ADAMIC_MUTANT='+mid+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+' > '+str(log)+' 2>&1',status=r.returncode,wall=time.monotonic()-start,binary_seconds=end[-1].get('Elapsed') if end else None,rows_run=ran,failed_rows=failed,cooked=any('test timed out' in e.get('Output','') for e in events))
  data.append(item);(p/'matrix.json').write_text(json.dumps(data,indent=2))
  if item['cooked']:break
