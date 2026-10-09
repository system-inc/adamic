import subprocess, pathlib, time, json, os
root=pathlib.Path('/workspace/adamic'); out=pathlib.Path('/tmp/defend-census')
plans=[('R1','internal/lower/invariance.go','node.Kind == ast.KindParenthesizedExpression || !l.isExpression(node)','node.Kind == ast.KindSpreadElement || !l.isExpression(node)'),('B1','internal/lower/census_small.go','\t\tof = ir.Boolean\n','\t\tof = ir.String\n')]
results=[]
for mid,file,old,new in plans:
 p=root/file; original=p.read_text(); assert original.count(old)==1
 start=time.monotonic()
 try:
  p.write_text(original.replace(old,new))
  with (out/(mid+'.diff')).open('w') as f: subprocess.run(['git','diff','--',file],cwd=root,stdout=f,check=True)
  with (out/(mid+'.vet.log')).open('w') as f: subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=f,stderr=subprocess.STDOUT,check=True)
  env=os.environ.copy(); env['ADAMIC_BUILD_CACHE_DIR']=str(out/'cache'/mid)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']
  with (out/(mid+'.log')).open('w') as f: run=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for line in (out/(mid+'.log')).read_text().splitlines():
   try: events.append(json.loads(line))
   except ValueError: pass
  failed=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']]
  passed=[e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']]
  skipped=[e['Test'] for e in events if e.get('Action')=='skip' and 'Test' in e and '/' not in e['Test']]
  evidence=[e['Output'].strip() for e in events if e.get('Action')=='output' and e.get('Test') in failed and ('.go:' in e.get('Output','') or 'FAIL' in e.get('Output',''))]
  r=dict(mutant=mid,file=file,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+' > '+mid+'.log 2>&1',exit=run.returncode,seconds=time.monotonic()-start,rows_failed=failed,rows_passed=passed,rows_skipped=skipped,evidence=evidence)
  results.append(r); (out/'matrix.json').write_text(json.dumps(results,indent=2)); print(mid,json.dumps({k:r[k] for k in ['exit','seconds','rows_failed','evidence']}),flush=True)
 finally: p.write_text(original)
