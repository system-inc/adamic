import pathlib,json,subprocess,os,time,difflib
r=pathlib.Path(__file__).resolve().parent;results=[]
for x in json.loads((r/'plan.json').read_text()):
 p=pathlib.Path(x['file']);s=p.read_text();assert s.count(x['old'])==1,(x['mutant'],s.count(x['old']));changed=s.replace(x['old'],x['new'])
 diff=''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p)));dp=r/(x['mutant']+'.diff');dp.write_text(diff)
 subprocess.run(['git','apply','--check',str(dp)],check=True);subprocess.run(['git','apply',str(dp)],check=True)
 try:
  with (r/(x['mutant']+'-vet.log')).open('w') as log:vet=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0,x['mutant']
  env=os.environ.copy();env.update(ADAMIC_BUILD_CACHE_DIR='/tmp/defend-check-pragmas/cache/'+x['mutant'])
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];start=time.monotonic()
  with (r/(x['mutant']+'.log')).open('w') as log:run=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
  ev=[]
  for line in (r/(x['mutant']+'.log')).read_text().splitlines():
   try:ev.append(json.loads(line))
   except:pass
  failed=sorted(set(e['Test'] for e in ev if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']));passed=sorted(set(e['Test'] for e in ev if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']))
  outputs=[e.get('Output','').strip() for e in ev if e.get('Test','').split('/')[0] in failed and e.get('Action')=='output' and ': ' in e.get('Output','')]
  res=dict(x,rows_failed=failed,rows_passed=passed,exit=run.returncode,wall_seconds=time.monotonic()-start,command=cmd,failing_output=outputs,panic=any('panic:' in e.get('Output','') for e in ev));results.append(res);(r/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(x['mutant'],failed,round(res['wall_seconds'],2),flush=True)
 finally:subprocess.run(['git','apply','-R',str(dp)],check=True)
