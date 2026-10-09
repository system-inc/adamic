import pathlib,subprocess,os,time,json
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-mediaquery'); results=[]
env=os.environ.copy();env['ADAMIC_MEDIA_QUERY_LIBRARY']='/tmp/u132/library'
for test in ['TestEachGapStandsWhereGapsMdSaysItDoes','TestThePortParsesAsGoCohereDoes']:
 for n in range(1,4):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/mediaquery/','-run','^'+test+'$'];start=time.monotonic();log=p/(test+'-'+str(n)+'.log')
  with log.open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',env=env,stdout=f,stderr=subprocess.STDOUT)
  sec=None
  for s in log.read_text().splitlines():
   try:x=json.loads(s)
   except:continue
   if x.get('Action')=='pass' and 'Test' not in x:sec=x['Elapsed']
  results.append(dict(test=test,trial=n,seconds=sec,wall=time.monotonic()-start,exit=r.returncode,command=cmd));(p/'timings.json').write_text(json.dumps(results,indent=2)+'\n');print(test,n,sec,r.returncode,flush=True)
