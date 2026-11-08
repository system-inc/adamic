from pathlib import Path
import json,hashlib,tarfile,collections
s=Path('/workspace/scratch/core-finish');e=s/'evidence';raw=s/'raw';summary=json.loads((e/'runtime-counts.json').read_text());plan=json.loads(Path('/workspace/adamic/stage3/lane/shard-plan.json').read_text());assert summary['complete'] and summary['main']=='2a95448b8fadbcd30549a148bd28da5f8b75f907';units=[];records=[];oracle=collections.Counter();processes=0;cast=0
for piece in summary['pieces']:
 archive=e/piece['archive'];assert hashlib.sha256(archive.read_bytes()).hexdigest()==piece['sha256'] and archive.stat().st_size<20_000_000
 with tarfile.open(archive) as tf:
  for line in tf.extractfile('SHA256SUMS').read().decode().splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(tf.extractfile(name).read()).hexdigest()==digest
  rows=[json.loads(l) for l in tf.extractfile('observations.jsonl')];assert sorted(r['unit'] for r in rows)==sorted(piece['units']);assert not piece['failures']
  for row in rows:
   assert row['childExit']==0 and row['report']['status']=='pass' and not row['report']['errors'];assert row['processes'];assert json.loads(next(x['text'] for x in row['raw'] if x['path']=='report.json'))==row['report'];oracle.update(row['report']['counts'])
   for process in row['processes']:
    assert process['counterexamples']==0 and process['calls']==len(process['records']);assert all(r['equal'] for r in process['records']);cast+=process['castFailureCalls'];processes+=1
    for r in process['records']:records.append({'unit':row['unit'],**r})
 units+=piece['units']
assert units==summary['units']==[u['name'] for u in plan['units']];assert len(units)==len(set(units))==398;assert len(records)==summary['totalEntries']==summary['counts']['deprecate']==6;assert cast==0;assert processes==summary['nodeProcesses'];assert dict(oracle)==summary['oracle']==plan['counts']
after=json.loads((raw/'moved-real-six-unit.json').read_text());assert after['passed'] and after['actualCalls']==6
def key(r):return json.dumps({'type':r['type'],'nameDescriptor':r['nameDescriptor'],'functionText':r['functionText'],'name':r['original']},sort_keys=True)
assert collections.Counter(key(r) for r in records)==collections.Counter(key(r) for r in after['records'])
classes=collections.Counter(r['type']+' / '+('own data name' if r['nameDescriptor'] and not r['nameDescriptor']['getter'] else 'own getter name' if r['nameDescriptor'] else 'no own name') for r in records);names=collections.Counter(r['original'] for r in records)
result={'passed':True,'acceptedUnitsExactlyOnce':398,'pieces':20,'processRecords':processes,'calls':6,'castFailureCalls':0,'counterexamples':0,'oracle':dict(oracle),'allArchiveAndInnerHashesVerified':True,'allRawReportsMatch':True,'nameAndInputMultisetIdenticalAcrossActualBeforeAndMovedCalls':True,'inputClasses':dict(classes),'returnedNames':dict(names),'unitsWithCalls':sorted({r['unit'] for r in records}),'sumPieceSeconds':sum(p['seconds'] for p in summary['pieces']),'largestPieceSeconds':max(p['seconds'] for p in summary['pieces']),'realCallRecords':records};(raw/'whole-lane-pair-reconciliation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k!='realCallRecords'}))
