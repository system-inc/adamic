from pathlib import Path
import collections,hashlib,json,sys,tarfile
out=Path(sys.argv[1]);root=Path('/workspace/adamic');plan=json.loads((root/'stage3/lane/shard-plan.json').read_text());summary=json.loads((out/'COUNT.json').read_text());byname={u['name']:u for u in plan['units']};units=[];counts=collections.Counter();oracle=collections.Counter();values=collections.Counter();hosts=collections.Counter();descriptors=collections.Counter();ready=0;pieces=[]
for archive in sorted(out.glob('piece-*.tar.gz')):
 digest=hashlib.sha256(archive.read_bytes()).hexdigest();assert (out/(archive.name+'.sha256')).read_text().split()[0]==digest
 with tarfile.open(archive,'r:gz') as tf:
  names=tf.getnames();base=archive.name.removesuffix('.tar.gz');record=json.load(tf.extractfile(base+'/piece.json'));children=json.load(tf.extractfile(base+'/children.json'));saved=json.load(tf.extractfile(base+'/events.json'));assert not record['failed'];assert record['frozenInputsUnchanged'];units+=record['units'];raw=[]
  assert [c['unit'] for c in children]==record['units']
  for child in children:
   assert child['child_exit']==0 and child['stop_reason'] is None
   report=json.load(tf.extractfile('results/'+child['unit']+'/report.json'));assert report==child['report'] and report['status']=='pass';unit=byname[child['unit']];assert report['counts']==unit['expect'];assert report['baseline_diffs']==unit['baseline_diffs'];assert report['upstream_exit']==(1 if unit['expect']['fail'] else 0);observed=json.load(tf.extractfile('results/'+child['unit']+'/observations.json'));identities=collections.Counter((r['task']['runner'],r['task']['file'],' '.join(p for p in r['name'] if p)) for r in observed);digest=hashlib.sha256(json.dumps(sorted((*k,v) for k,v in identities.items()),sort_keys=True,separators=(',',':')).encode()).hexdigest();assert digest==unit['names_digest'];assert dict(collections.Counter(r['status'] for r in observed))=={k:v for k,v in unit['expect'].items() if v};oracle.update(report['counts'])
  for item in saved:
   rawpath=base+'/calls/'+item['unit']+'/'+item['file'];rows=[json.loads(l) for l in tf.extractfile(rawpath).read().decode().splitlines()];assert rows==item['rows'];raw+=rows
  for row in raw:
   if row['event']=='observer-ready':ready+=1
   if row['event']=='evaluation':
    actual=row['actual'];kind=actual['type'];value=actual.get('value');expected=False if kind in ['undefined','null'] else bool(value) if kind in ['boolean','string'] else value not in ['0','-0','NaN'] if kind in ['number','bigint'] else True if kind in ['object','function','symbol'] else None
    assert expected is not None and row['truthy']==expected
    hosts[str(row['nodeProcess'])]+=1;descriptors[str(row['ownProperty'])]+=1
    counts['evaluations']+=1;counts['truthy' if row['truthy'] else 'falsy']+=1;values[json.dumps(row['actual'],sort_keys=True)]+=1
  pieces.append(dict(archive=archive.name,sha256=digest,units=len(record['units'])))
assert units==[u['name'] for u in plan['units']]
assert len(set(units))==len(units)==398
assert dict(oracle)==plan['counts']==summary['oracleCounts']
assert dict(values)==summary['actualValues']
for k in ['evaluations','truthy','falsy']:assert counts[k]==summary[k]
assert summary['complete']
result=dict(independentReconciliation='pass',units=len(units),pieces=pieces,counts=dict(counts),actualValues=dict(values),oracleCounts=dict(oracle),nodeProcessIdentity=dict(hosts),ownProperty=dict(descriptors),observerReadyEvents=ready,rawRowsReopened=True,ordinaryShardReportsReopened=True,allArchiveHashesVerified=True)
(out/'reconciliation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
