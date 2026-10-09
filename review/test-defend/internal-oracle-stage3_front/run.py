import json,subprocess,os,time
from pathlib import Path
root=Path('/workspace/adamic');s=Path('/tmp/defend-typed');plans=json.loads((s/'plan.json').read_text()); originals={e['file']:(root/e['file']).read_text() for p in plans for e in p['edits']}
pattern='^(TestTypedArray.*|TestStage3Enum.*|TestNativeAgreesWithNode)$/^(internal|stage3)$/^(oracle|fixtures)$/^(testdata|enums)$/^(typed_arrays_.*|writes_past_end.a|write_after_shrink.a|[0-9].*)$'
def run(id):
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=str(s/'cache'/('corrected-'+id));env['ADAMIC_GATE_UNCACHED']='1'
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern];start=time.monotonic()
 with (s/(id+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 events=[]
 for l in (s/(id+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 statuses={e['Test']:e['Action'] for e in events if e.get('Test') and e['Action'] in ['pass','fail','skip']}
 data={'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'uncached':True,'wall_seconds':time.monotonic()-start,'exit':r.returncode,'binary_seconds':next((e.get('Elapsed') for e in reversed(events) if 'Test' not in e and e['Action'] in ['pass','fail']),None),'statuses':statuses,'failing_output':[e['Output'].strip() for e in events if e.get('Test') and e.get('Output') and ('test.go:' in e['Output'])]};(s/(id+'.json')).write_text(json.dumps(data,indent=2));print(id,data['exit'],data['binary_seconds'],{k:v for k,v in statuses.items() if v!='pass'},flush=True);return data
try:
 baseline=run('narrow-baseline')
 if baseline['exit']:raise RuntimeError('clean narrowed baseline not green')
 for p in plans:
  for e in p['edits']:(root/e['file']).write_text(originals[e['file']].replace(e['before'],e['after'],1))
  (s/(p['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',*[e['file'] for e in p['edits']]],cwd=root))
  with (s/(p['mutant']+'-vet.log')).open('w') as out:subprocess.run(['go','vet','./internal/native/','./internal/javascript/'],cwd=root,stdout=out,stderr=subprocess.STDOUT,check=True)
  run(p['mutant'])
  for name,t in originals.items():(root/name).write_text(t)
finally:
 for name,t in originals.items():(root/name).write_text(t)
