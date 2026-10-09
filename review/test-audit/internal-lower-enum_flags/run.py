import json,time,subprocess,os,re
from pathlib import Path
P=Path('review/test-audit/internal-lower-enum_flags');S=Path('/tmp/u031');plan=json.loads((P/'plan.json').read_text()); rows=json.loads((S/'rows.json').read_text()); originals={str(Path(str(f).replace('/tmp/u031/original/','')).resolve()):str(f.resolve()) for f in (S/'original').rglob('*.go')}; empty=S/'empty.go';empty.write_text('package lower\n'); originals[str(Path('internal/lower/audit_mutant.go').resolve())]=str(empty)
validation=[]
for m in plan:
 prior=next((x for x in json.loads((P/'validation-initial.json').read_text()) if x['id']==m['id'] and x['exit']==0),None)
 if prior:
  validation.append(prior);continue
 folder=S/'standalone'/m['id'];overlay=dict(originals);overlay[str(Path(m['file']).resolve())]=str((folder/Path(m['file']).name).resolve());(folder/'overlay.json').write_text(json.dumps({'Replace':overlay}));cmd=['timeout','90','go','vet','-overlay',str(folder/'overlay.json'),'./internal/lower/'];start=time.monotonic()
 with open(S/(m['id']+'-vet.log'),'w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 validation.append(dict(id=m['id'],command=' '.join(cmd),exit=r.returncode,seconds=time.monotonic()-start,output=(S/(m['id']+'-vet.log')).read_text()))
 print('vet',m['id'],r.returncode,flush=True)
(P/'validation.json').write_text(json.dumps(validation,indent=2));assert all(x['exit']==0 for x in validation)
def run(id,pattern,label):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',pattern];env=dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u031/cache/'+id);start=time.monotonic();log=S/(label+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for line in log.read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 statuses={e['Test']:e['Action'] for e in events if e.get('Test') and e['Action'] in ['pass','fail','skip']};binary=next((e.get('Elapsed') for e in reversed(events) if not e.get('Test') and e['Action'] in ['pass','fail']),None);aborted=r.returncode==124 or any('panic:' in e.get('Output','') for e in events)
 record=dict(id=id,label=label,command="ADAMIC_MUTANT='"+id+"' ADAMIC_BUILD_CACHE_DIR='/tmp/u031/cache/"+id+"' "+' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,binary_seconds=binary,status=statuses,aborted=aborted,events=events);(P/(label+'.json')).write_text(json.dumps(record,indent=2));return record
matrix=[]
for m in plan:
 d=run(m['id'],'.',m['id']); reruns=[]
 if d['aborted']:
  for row in rows:reruns.append(run(m['id'],'^'+row+'$',m['id']+'-'+row))
 matrix.append(dict(id=m['id'],aborted=d['aborted'],failed=[k for k,v in d['status'].items() if '/' not in k and v=='fail'],reruns=[r['label'] for r in reruns]));(P/'matrix-index.json').write_text(json.dumps(matrix,indent=2));print('matrix',matrix[-1],flush=True)
