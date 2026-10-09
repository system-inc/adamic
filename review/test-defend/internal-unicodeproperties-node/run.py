import json,subprocess,time,os
from pathlib import Path
root=Path('/workspace/adamic');s=Path('/tmp/defend-unicode'); f=root/'internal/unicodeproperties/tables.go';original=f.read_text()
rows=['TestVersionMatchesNode','TestNodeAgrees','TestNodeStringProperties','TestBinaryAliases','TestGeneralCategoryAliases','TestScriptAliasSet','TestRejectedNames','TestStringPropertyCensus','TestSetBoundaries','TestKnownMembership']
pattern='^('+'|'.join(rows)+')$';(s/'matrix-rows.json').write_text(json.dumps(rows,indent=2))
def run(id):
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(s/'cache'/id)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/unicodeproperties/','-run',pattern]
 start=time.monotonic()
 with (s/(id+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 events=[]
 for line in (s/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 statuses={e['Test']:e['Action'] for e in events if e.get('Test') in rows and e['Action'] in ['pass','fail','skip']}
 info={'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'wall_seconds':time.monotonic()-start,'exit':r.returncode,'binary_seconds':next((e.get('Elapsed') for e in reversed(events) if 'Test' not in e and e['Action'] in ['pass','fail']),None),'statuses':statuses,'failures':[e['Output'].strip() for e in events if e.get('Test') in rows and 'Output' in e and (': ' in e['Output']) and not e['Output'].startswith('===')]}
 (s/(id+'.json')).write_text(json.dumps(info,indent=2));print(id,r.returncode,info['binary_seconds'],statuses,flush=True);return info
try:
 baseline=run('narrow-baseline')
 if baseline['exit']:raise RuntimeError('red narrowed baseline')
 for p in json.loads((s/'plan.json').read_text()):
  pos=p['position']; mutated=original[:pos]+p['after']+original[pos+len(p['before']):];f.write_text(mutated)
  (s/(p['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',p['file']],cwd=root))
  with (s/(p['mutant']+'-vet.log')).open('w') as out:
   subprocess.run(['go','vet','./internal/unicodeproperties/'],cwd=root,stdout=out,stderr=subprocess.STDOUT,check=True)
  run(p['mutant']);f.write_text(original)
finally:f.write_text(original)
