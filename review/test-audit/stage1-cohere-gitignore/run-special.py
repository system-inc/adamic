import pathlib,subprocess,time,json,os
p=pathlib.Path('review/test-audit/stage1-cohere-gitignore');groups=json.loads((p/'groups.json').read_text());records=[]
def run(name,cmd,env=None):
 start=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(cmd,env=dict(os.environ,**(env or {})),stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for line in (p/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 rec=dict(name=name,command=cmd,env=env or {},wall=time.monotonic()-start,exit=r.returncode,failures=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']],skips=[e['Test'] for e in events if e.get('Action')=='skip' and 'Test' in e],terminal=[e for e in events if e.get('Action') in ('pass','fail') and 'Test' not in e]);records.append(rec);(p/'special-runs.json').write_text(json.dumps(records,indent=2));print(name,rec['exit'],round(rec['wall'],3),rec['failures'],flush=True);return rec
for id in ['W1','S1','P1','P2','P3']:
 diff=str(p/'diffs'/f'{id}.diff');subprocess.run(['git','apply','--check',diff],check=True);subprocess.run(['git','apply',diff],check=True)
 env={'COHERE_GIT_SOURCE':'/tmp/u094/git-source','ADAMIC_BUILD_CACHE_DIR':'/tmp/u094/cache/'+id}
 try:
  if id in ['W1','S1','P3']:
   pkg='./internal/lower/' if id=='P3' else './stage1/cohere/gitignore/'
   if run(id+'-vet',['timeout','90','go','vet',pkg])['exit']:raise RuntimeError('vet failed')
  todo={'W1':['witnesses','path','planted'],'S1':['growth'],'P1':['answers'],'P2':['path'],'P3':['gaps']}[id]
  if id=='P3':env['ADAMIC_EMPTY_PROBE']='P3'
  if id in ['P1','P2','P3']:
   target={'P1':'main.ts','P2':'pathsweep.ts','P3':'gaps/6_boolean_element.ts'}[id]
   if run(id+'-build',['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/gitignore/'+target,'-o','/tmp/u094/'+id+'-port','--sanitize'],env)['exit']:raise RuntimeError('probe build failed')
  for name in todo:run(id+'-'+name,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/gitignore/','-run',groups[name]],env)
  if id=='W1':run('W1-largest',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/gitignore/','-run','^TestThePortAnswersAsGoCohereAndGitDo_026$'],dict(env,ADAMIC_GITIGNORE_LARGEST='1'))
 finally:subprocess.run(['git','apply','-R',diff],check=True)
for i in [2,3]:run('largest-timing-'+str(i),['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/gitignore/','-run','^TestThePortAnswersAsGoCohereAndGitDo_026$'],{'COHERE_GIT_SOURCE':'/tmp/u094/git-source','ADAMIC_GITIGNORE_LARGEST':'1'})
