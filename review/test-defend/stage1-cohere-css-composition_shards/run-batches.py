import pathlib,json,subprocess,os,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-css-composition_shards';results=json.loads((p/'results.json').read_text());source=root/'stage1/cohere/css/main.ts';base=source.read_text();m=next(x for x in json.loads((p/'plan.json').read_text()) if x['mutant']=='D1')
def run(ident,i,names):
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-css/cache/'+ident
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^('+'|'.join(names)+')$'];log=ident+'-parser-batch-'+str(i)+'.log';start=time.monotonic()
 with (p/log).open('w') as out:q=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 ev=[]
 for s in (p/log).read_text().splitlines():
  try:ev.append(json.loads(s))
  except:pass
 rr=dict(mutant=ident,group='parser-batch-'+str(i),command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],wall=round(time.monotonic()-start,3),exit=q.returncode,failed=[x['Test'] for x in ev if x['Action']=='fail' and x.get('Test')],passed=[x['Test'] for x in ev if x['Action']=='pass' and x.get('Test')],skipped=[x['Test'] for x in ev if x['Action']=='skip' and x.get('Test')],cooked=any('test timed out' in x.get('Output','') for x in ev),outputs=[])
 results.append(rr);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(ident,i,rr['wall'],q.returncode,len(rr['passed']),rr['failed'],flush=True)
 if q.returncode:raise SystemExit('batch failed, do not claim completed matrix')
try:
 for ident in ['clean','D1']:
  source.write_text(base if ident=='clean' else base.replace(m['old'],m['new'],1))
  for i,start in enumerate(range(0,388,100)):run(ident,i,['TestThePortParsesAsGoCohereDoes_%03d'%n for n in range(start,min(388,start+100))])
finally:source.write_text(base)
