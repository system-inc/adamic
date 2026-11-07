from pathlib import Path
import collections,hashlib,json,sys,tarfile
out=Path(sys.argv[1]);root=Path('/workspace/adamic');plan=json.loads((root/'stage3/lane/shard-plan.json').read_text());summary=json.loads((out/'COUNT.json').read_text());units=[];counts=collections.Counter();oracle=collections.Counter();values=collections.Counter();pieces=[]
for archive in sorted(out.glob('piece-*.tar.gz')):
 digest=hashlib.sha256(archive.read_bytes()).hexdigest();assert (out/(archive.name+'.sha256')).read_text().split()[0]==digest
 with tarfile.open(archive,'r:gz') as tf:
  names=tf.getnames();base=archive.name.removesuffix('.tar.gz');record=json.load(tf.extractfile(base+'/piece.json'));children=json.load(tf.extractfile(base+'/children.json'));saved=json.load(tf.extractfile(base+'/events.json'));assert not record['failed'];assert record['frozenInputsUnchanged'];units+=record['units'];raw=[]
  assert [c['unit'] for c in children]==record['units']
  for child in children:
   assert child['child_exit']==0 and child['stop_reason'] is None
   report=json.load(tf.extractfile('results/'+child['unit']+'/report.json'));assert report==child['report'] and report['status']=='pass';oracle.update(report['counts'])
  for item in saved:
   rawpath=base+'/calls/'+item['unit']+'/'+item['file'];rows=[json.loads(l) for l in tf.extractfile(rawpath).read().decode().splitlines()];assert rows==item['rows'];raw+=rows
  for row in raw:
   if row['event']=='evaluation':
    counts['evaluations']+=1;counts['truthy' if row['truthy'] else 'falsy']+=1;values[json.dumps(row['actual'],sort_keys=True)]+=1
  pieces.append(dict(archive=archive.name,sha256=digest,units=len(record['units'])))
assert units==[u['name'] for u in plan['units']]
assert len(set(units))==len(units)==398
assert dict(oracle)==plan['counts']==summary['oracleCounts']
assert dict(values)==summary['actualValues']
for k in ['evaluations','truthy','falsy']:assert counts[k]==summary[k]
assert summary['complete']
result=dict(independentReconciliation='pass',units=len(units),pieces=pieces,counts=dict(counts),actualValues=dict(values),oracleCounts=dict(oracle),rawRowsReopened=True,ordinaryShardReportsReopened=True,allArchiveHashesVerified=True)
(out/'reconciliation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
