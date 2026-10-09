from pathlib import Path
import json,subprocess,os,time
root=Path('/workspace/adamic');s=Path('/tmp/defend-regexp');plans=json.loads((s/'plan.json').read_text())
for p in plans:
 f=root/p['file'];original=f.read_text()
 try:
  assert original.count(p['before'])==1
  f.write_text(original.replace(p['before'],p['after'],1));(s/(p['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',p['file']],cwd=root))
  with (s/(p['mutant']+'-vet.log')).open('w') as out:subprocess.run(['go','vet','./internal/regexp/'],cwd=root,stdout=out,stderr=subprocess.STDOUT,check=True)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(s/'cache'/p['mutant'])
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','.'];start=time.monotonic()
  with (s/(p['mutant']+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
  events=[]
  for line in (s/(p['mutant']+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  statuses={e['Test']:e['Action'] for e in events if 'Test' in e and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']}
  data={'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'wall_seconds':time.monotonic()-start,'exit':r.returncode,'binary_seconds':next((e.get('Elapsed') for e in reversed(events) if 'Test' not in e and e['Action'] in ['pass','fail']),None),'statuses':statuses,'failures':[e['Output'].strip() for e in events if e.get('Test')=='TestNodeAgreement' and 'Output' in e and ('oracle_test.go:' in e['Output'] or 'parser=true node=false' in e['Output'])]}
  (s/(p['mutant']+'.json')).write_text(json.dumps(data,indent=2));print(json.dumps(data),flush=True)
 finally:f.write_text(original)
