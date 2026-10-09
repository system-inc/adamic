import pathlib,json,os,time,subprocess
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-mediaquery';plan=json.loads((out/'plan.json').read_text());records=[]
for m in plan:
 diff=out/'diffs'/(m['id']+'.diff');subprocess.run(['git','apply',str(diff)],cwd=root,check=True)
 try:
  env=os.environ.copy();env['PATH']='/tmp/u132/bin:'+env['PATH'];env['ADAMIC_AUDIT_ID']=m['id'];env['ADAMIC_MEDIA_QUERY_LIBRARY']='/tmp/u132/library';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u132/cache/'+m['id']
  start=time.monotonic()
  if m['file'].endswith('.go'):
   pkg='./internal/lower/' if m['file'].startswith('internal/lower') else './stage1/cohere/mediaquery/'
   with (out/(m['id']+'-vet.log')).open('w') as f:v=subprocess.run(['timeout','90','go','vet',pkg],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   assert v.returncode==0,m['id']
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/mediaquery/','-run','.']
  with (out/(m['id']+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  records.append(dict(id=m['id'],wall=time.monotonic()-start,exit=r.returncode,command=cmd))
  if 'panic:' in (out/(m['id']+'.log')).read_text():
   for test in ['TestEachGapStandsWhereGapsMdSaysItDoes','TestThePortParsesAsGoCohereDoes']:
    alone=cmd[:-1]+['^'+test+'$'];a=time.monotonic()
    with (out/(m['id']+'-'+test+'.log')).open('w') as f:rr=subprocess.run(alone,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
    records.append(dict(id=m['id'],test=test,wall=time.monotonic()-a,exit=rr.returncode,command=alone))
  (out/'runs.json').write_text(json.dumps(records,indent=2)+'\n');print(m['id'],r.returncode,records[-1]['wall'],flush=True)
 finally:subprocess.run(['git','restore','--source=HEAD','--',m['file']],cwd=root,check=True)
