import pathlib,subprocess,json,time,os,sys
p=pathlib.Path('/tmp/defend-yaml-gaps');plan=json.loads((p/'plan.json').read_text());runs=[]
for change in plan:
 mid=change['mutant']
 if len(sys.argv)>1 and mid not in sys.argv[1:]:continue
 diff=p/(mid+'.diff');subprocess.run(['git','apply','--check',str(diff)],check=True)
 subprocess.run(['git','apply',str(diff)],check=True)
 try:
  env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']=str(p/'library');env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/mid)
  if mid in ['D1','D2']:
   with (p/(mid+'.vet.log')).open('w') as f:subprocess.run(['go','vet','./internal/lower/'],stdout=f,stderr=subprocess.STDOUT,env=env,check=True)
   subprocess.run(['python3',str(p/'run.py'),mid],check=True,env=env)
  else:
   rows=['TestLexerGaps','TestStructuralPositionRefusal','TestLexerMatchesGo','TestPropsMatchGo','TestCSTMatchesGo']
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+'|'.join(rows)+')$']
   start=time.monotonic()
   with (p/(mid+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
   result=dict(id=mid,rows=rows,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-start);runs.append(result)
   (p/'port-runs.json').write_text(json.dumps(runs,indent=2)+'\n');print(mid,r.returncode,round(result['wall_seconds'],3),flush=True)
 finally:
  subprocess.run(['git','apply','--reverse',str(diff)],check=True)
