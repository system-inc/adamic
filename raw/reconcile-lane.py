import collections,hashlib,json
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/meta-spread-lane-final');raw=Path(__file__).parent
plan=json.loads((repo/'stage3/lane/shards.json').read_text())
ledger=json.loads((root/'ledger.json').read_text());assert len(ledger['units'])==len(plan['units'])==398
assert all(row['exit']==0 and row['status']=='pass' and row['seconds']<30 for row in ledger['units'])
rows=[json.loads(line) for line in (raw/'lane-events.jsonl').read_text().splitlines()]
loads=[row for row in rows if row['event']=='load'];starts=[row for row in rows if row['event']=='start'];metas=[row for row in rows if row['event']=='meta']
keys=['cat','ph','ts','pid','tid']
for row in metas:
 for field in ['keys','ownKeys','enumerableKeys']:assert row[field]==keys
 assert row['values']['cat']=='__metadata' and row['values']['ph']=='M' and row['values']['pid']==row['values']['tid']==1
assert {r['run'] for r in starts}=={r['run'] for r in metas},'a start failed before meta was built'
counts=collections.Counter();observations=0;covered=[]
expected_hash=hashlib.sha256(Path('/workspace/scratch/meta-spread-counting-final/built/local/run.js').read_bytes()).hexdigest()
for unit in plan['units']:
 name=unit['name'];report=json.loads((root/name/'report.json').read_text());observed=json.loads((root/name/'observations.json').read_text())
 assert report['counts']==unit['expect'] and report['status']=='pass'
 assert report['artifact']=='stage3-adapted-tree'
 counts.update(report['counts']);observations+=len(observed)
 harness=root/name/'adapted-tree/built/local/run.js'
 assert hashlib.sha256(harness.read_bytes()).hexdigest()==expected_hash
 assert any(row['file']==str(harness) for row in loads),'counter was not loaded in '+name
 covered.append(name)
expected=collections.Counter()
for unit in plan['units']:expected.update(unit['expect'])
assert dict(counts)==dict(expected)
platform=json.loads((repo/'stage3/lane/expected.json').read_text())['platforms']['linux/x64/v24.19.0']
assert dict(counts)==dict(zip(('pass','fail','skip'),(platform['passing'],platform['failing'],platform['pending'])))
control=json.loads((raw/'final-lane-control.json').read_text());assert control['passed'] and control['starts']==control['metaObservations']==1
result=dict(passed=True,route='STAGE3_ADAPTED_TREE with private counting bundles; unchanged measure_shards.py and shard.py',shards=398,counts=dict(counts),observations=observations,startTracingRuns=len(starts),actualMetaObservations=len(metas),actualMetaKeysAndOrder=[r['keys'] for r in metas],expectedMetaKeys=keys,otherKeySets=0,bundleLoads=len(loads),instrumentedProcesses=len({r['pid'] for r in loads}),all398HarnessCopiesHashIdentical=True,all398ShardsLoadedCounter=True,counterHarnessSha256=expected_hash,controlStarts=1,controlMetaObservations=1,controlActualKeys=control['keys'],seconds=ledger['wall_seconds'],maximumShardSeconds=ledger['maximum'],concurrency=ledger['concurrency'])
(raw/'lane-summary.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
