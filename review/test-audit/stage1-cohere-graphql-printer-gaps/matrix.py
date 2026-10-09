from pathlib import Path
import subprocess,time,json,os
p=Path('review/test-audit/stage1-cohere-graphql-printer-gaps');rows=json.loads((p/'scope.json').read_text())['rows'];data=[]
for mid in ['control','M01','M02','M03','P01']:
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/workspace/u097-products/'+mid
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run','^('+'|'.join(rows)+')$'];s=time.monotonic()
 with (p/(mid+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 e=[]
 for l in open(p/(mid+'.log')):
  try:e.append(json.loads(l))
  except:pass
 end=[x for x in e if x.get('Action') in ['pass','fail'] and not x.get('Test')]
 data.append(dict(id=mid,status=r.returncode,command=cmd,env={'ADAMIC_MUTANT':mid,'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},wall=time.monotonic()-s,seconds=end[-1].get('Elapsed') if end else None,rows_run=sorted(set(x['Test'].split('/')[0] for x in e if x.get('Action')=='run' and x.get('Test'))),failed_rows=sorted(set(x['Test'].split('/')[0] for x in e if x.get('Action')=='fail' and x.get('Test'))),cooked=any('test timed out' in x.get('Output','') for x in e)))
 (p/'matrix.json').write_text(json.dumps(data,indent=2))
 if data[-1]['cooked']:raise SystemExit(mid+' cooked')
 if mid=='control' and r.returncode:raise SystemExit('inactive selector red')
