import pathlib,subprocess,time,json,os,re
p=pathlib.Path('review/test-audit/stage1-cohere-gitignore'); plan=json.loads((p/'plan.json').read_text()); groups=json.loads((p/'groups.json').read_text()); records=json.loads((p/'matrix-runs.json').read_text()) if (p/'matrix-runs.json').exists() else []
lines=(p/'lower-functions.log').read_text().splitlines(); reached=[s for s in lines if float(s.split()[-1].rstrip('%'))>0]; (p/'reached-lower-functions.txt').write_text('\n'.join(reached)+'\n')
def run(name,cmd,env=None):
 start=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(cmd,env=dict(os.environ,**(env or {})),stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for line in (p/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 rec=dict(name=name,command=cmd,env=env or {},wall=time.monotonic()-start,exit=r.returncode,failures=[e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']],skips=[e['Test'] for e in events if e.get('Action')=='skip' and 'Test' in e],terminal=[e for e in events if e.get('Action') in ('pass','fail') and 'Test' not in e]);records.append(rec);(p/'matrix-runs.json').write_text(json.dumps(records,indent=2));print(name,rec['exit'],round(rec['wall'],3),rec['failures'],flush=True);return rec
for m in plan:
 diff=str(p/'diffs'/f"{m['id']}.diff"); subprocess.run(['git','apply','--check',diff],check=True);subprocess.run(['git','apply',diff],check=True)
 env={'COHERE_GIT_SOURCE':'/tmp/u094/git-source','ADAMIC_BUILD_CACHE_DIR':'/tmp/u094/cache/'+m['id']}
 try:
  if m['id']=='M4':
   if run('M4-vet',['timeout','90','go','vet','./internal/lower/'])['exit']:raise RuntimeError('vet failed')
  targets=['main.ts'] if m['id']!='M4' else ['gaps/6_boolean_element.ts','main.ts']
  if m['id']=='M1':targets.append('pathsweep.ts')
  for target in ([] if m['id']=='M1' else targets):
   cmd=['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/gitignore/'+target,'-o','/tmp/u094/'+m['id']+'-'+target.replace('/','-')+'-port','--sanitize']
   if run(m['id']+'-build-'+target.replace('/','-'),cmd,env)['exit']:raise RuntimeError('standalone native build failed')
  for name,regex in groups.items():
   if name=='witnesses':continue
   if m['id']=='M1' and name=='answers':regex='^TestThePortAnswersAsGoCohereAndGitDo_014$'
   if m['id']=='M1' and name=='answers':name='answers-narrow'
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/gitignore/','-run',regex]
   rec=run(m['id']+'-'+name,cmd,env)
   text=(p/(rec['name']+'.log')).read_text()
   if 'panic:' in text or 'cooked:' in text or rec['exit']==124:
    members=[s for s in (p/'list.log').read_text().splitlines() if s.startswith('Test') and re.search(regex,s)]
    for member in []:run(m['id']+'-solo-'+member,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/gitignore/','-run','^'+member+'$'],env)
 finally:subprocess.run(['git','apply','-R',diff],check=True)
