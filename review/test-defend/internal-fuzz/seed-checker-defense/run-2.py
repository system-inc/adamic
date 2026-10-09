import pathlib,subprocess,json,time,os
out=pathlib.Path('/tmp/defend-fuzz');file='internal/fuzz/generate.go';p=pathlib.Path(file)
plans=[('G2','TestGeneratedProgramsCheckAndLower','drop statement','\t\t\tleft = g.leafVariable(String, left)\n',''),('G3','TestGeneratedProgramsCheckAndLower','change constant','return compose(StringArray, "@e.map(("+item+") => @e)", g.expression(NumberArray, next), body)','return compose(StringArray, "@e.map(("+item+") => @e)", g.expression(StringArray, next), body)')]
source=p.read_text()
manifest=[dict(mutant=m,target=t,menu=menu,file=file,line=source[:source.index(old)].count('\n')+1,old=old,new=new)for m,t,menu,old,new in plans]
(out/'plan-2.json').write_text(json.dumps(manifest,indent=2)+'\n');results=json.loads((out/'matrix.json').read_text())
for item in manifest:
 mid=item['mutant'];original=p.read_text();assert original.count(item['old'])==1
 start=time.monotonic()
 try:
  p.write_text(original.replace(item['old'],item['new']))
  with(out/(mid+'.diff')).open('w')as f:subprocess.run(['git','diff','--',file],stdout=f,check=True)
  with(out/(mid+'.vet.log')).open('w')as f:subprocess.run(['go','vet','./internal/fuzz/'],stdout=f,stderr=subprocess.STDOUT,check=True)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(out/'cache'/mid)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/fuzz/','-run','.']
  with(out/(mid+'.log')).open('w')as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for line in(out/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except ValueError:pass
  failed=sorted(set(e['Test'].split('/')[0]for e in events if e.get('Action')=='fail'and'Test'in e))
  passed=[e['Test']for e in events if e.get('Action')=='pass'and'Test'in e and'/'not in e['Test']]
  panics=[e.get('Test','unknown')for e in events if'panic:'in e.get('Output','')]
  evidence=[e['Output'].strip()for e in events if e.get('OutputType')=='error'and e.get('Test','').split('/')[0]in failed]
  result=dict(**item,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+' > '+mid+'.log 2>&1',exit=r.returncode,wall_seconds=time.monotonic()-start,rows_failed=failed,rows_passed=passed,panics=panics,evidence=evidence)
  results.append(result);(out/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(mid,json.dumps(dict(exit=r.returncode,failed=failed,panics=panics,evidence=evidence[:2])),flush=True)
 finally:p.write_text(original)
