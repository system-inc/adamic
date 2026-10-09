import subprocess,os,time,json
from pathlib import Path
names=['TestProduct_YamlScalarsLowered','TestProduct_YamlScalarsNative','TestProduct_YamlScalarsGo','TestScalarsMatchGo family','TestScalarsMatchGoPlantedFailure','TestSchemaMatchesGo','TestSchemaMutants','TestSpeedCostProbes','TestUnistMatchesGo','TestUnistMutants','TestWidthsMatchGo']
env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']='/tmp/u153/library'
results=[]
for name in names:
 regex='^TestScalarsMatchGo(Union|_[0-9]{3})$' if name.endswith('family') else '^'+name+'$'
 for trial in range(3):
  path=Path('/tmp/u153')/('timing-'+name.replace(' ','-')+'-'+str(trial)+'.log')
  start=time.monotonic()
  with path.open('w') as f: r=subprocess.run(['timeout','95','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',regex],env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in path.read_text().splitlines():
   try: events.append(json.loads(l))
   except: pass
  sec=next((e.get('Elapsed') for e in reversed(events) if 'Test' not in e and e.get('Action') in ['pass','fail']),None)
  results.append(dict(test=name,trial=trial,seconds=sec,wall=time.monotonic()-start,status=r.returncode,path=str(path)))
  Path('/tmp/u153/timings.json').write_text(json.dumps(results,indent=2))
  print(name,trial,sec,r.returncode,flush=True)
  if r.returncode: break
