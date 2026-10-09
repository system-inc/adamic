import pathlib,subprocess,json,time,os
out=pathlib.Path('/tmp/defend-fuzz');file='internal/fuzz/generate.go';p=pathlib.Path(file)
plans=[
('G1','TestGeneratedProgramsCheckAndLower','change constant','return compose(StringArray, "@e.slice(@e)", g.expression(StringArray, next), g.expression(Number, next))','return compose(StringArray, "@e.sliceMissing(@e)", g.expression(StringArray, next), g.expression(Number, next))'),
('S1','TestOneSeedOneProgram','change constant','return GenerateWithout(seed, nil)','return GenerateWithout(1, nil)'),
('S2','TestOneSeedOneProgram','change constant','seed: seed}','seed: 0}'),
('S3','TestOneSeedOneProgram','swap arguments','rand.NewPCG(seed, 0x61646d6963)','rand.NewPCG(0x61646d6963, seed)'),
('R1','TestRegexProgramsPassTheChecker','drop statement','\t\tadd(statement("const finder = /a/g;"))\n',''),
('R2','TestRegexProgramsPassTheChecker','change constant','statement("const found = /a/g.exec(@e);", g.bounded(g.expression(String, 2)))','statement("const found = /a/g.execMissing(@e);", g.bounded(g.expression(String, 2)))'),
('R3','TestRegexProgramsPassTheChecker','change constant','compose(String, "@e.replace("+g.pick("/a/g", "/(a)/g", "/a/i")+", "+g.pick("\'[ $&]\'", "\'$1\'", "\'$$\'", "\'-\'")+" )", g.bounded(g.expression(String, 2)))','')]
# Freeze the exact first replacement line instead of approximating its quoting.
source=p.read_text();line=next(l for l in source.splitlines()if 'compose(String, "@e.replace("+g.pick' in l)
plans[-1]=('R3','TestRegexProgramsPassTheChecker','change constant',line,line.replace('g.expression(String, 2)','g.expression(Number, 2)'))
manifest=[dict(mutant=m,target=t,menu=menu,file=file,line=source[:source.index(old)].count('\n')+1,old=old,new=new)for m,t,menu,old,new in plans]
(out/'plan.json').write_text(json.dumps(manifest,indent=2)+'\n');results=[]
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
