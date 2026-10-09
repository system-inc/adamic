import pathlib,json,subprocess,os,time,re
p=pathlib.Path('/tmp/u099/evidence');scope=json.loads((p/'scope.json').read_text());env=os.environ.copy();env['ADAMIC_GRAPHQL_PRETTIER']='/tmp/u099-prettier'
for group in scope['groups']:
 pattern='^('+'|'.join(group['members'])+')$'
 for repeat in range(1,4):
  id=group['test'].replace(' ','_')+'-time'+str(repeat);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run',pattern];start=time.monotonic()
  with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/(id+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  end=next((e for e in reversed(events) if e.get('Action') in ['pass','fail'] and 'Test' not in e),{});meta={'row':group['test'],'command':'ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier '+__import__('shlex').join(cmd),'exit':r.returncode,'wall':time.monotonic()-start,'seconds':end.get('Elapsed'),'timeout':any('test timed out' in e.get('Output','') for e in events)};(p/(id+'-run.json')).write_text(json.dumps(meta,indent=2));print(id,meta['seconds'],r.returncode,flush=True)
  if r.returncode!=0 and not meta['timeout']:raise SystemExit('Red timing baseline, stop')
