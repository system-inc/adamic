"""Reopen all raw pieces and count helper entries independently of the controller summary."""
from pathlib import Path
import json,hashlib,tarfile,collections
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/getfunctionname');e=s/'evidence';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());summary=json.loads((e/'runtime-counts.json').read_text());assert summary['complete'];refs=json.loads((s/'raw/references.json').read_text());ids=[f"{r['file']}:{r['line']}:{r['column']}" for r in refs['callSites']];counts=collections.Counter(dict.fromkeys(ids,0));attempts=collections.Counter(dict.fromkeys(ids,0));oracle=collections.Counter();units=[];unknown={};total=0;processes=0;executedUnits=[]
for piece in summary['pieces']:
 archive=e/piece['archive'];assert hashlib.sha256(archive.read_bytes()).hexdigest()==piece['sha256'];assert not piece['failures'];assert piece['seconds']<1490
 with tarfile.open(archive) as tf:
  for line in tf.extractfile('SHA256SUMS').read().decode().splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(tf.extractfile(name).read()).hexdigest()==digest
  metadata=json.loads(tf.extractfile('piece.json').read());assert all(metadata[k]==piece[k] for k in metadata)
  rows=[json.loads(l) for l in tf.extractfile('observations.jsonl')];assert sorted(r['unit'] for r in rows)==sorted(piece['units'])
  for row in rows:
   assert row['childExit']==0 and row['report']['status']=='pass' and not row['report']['errors'];assert row['processes'];assert json.loads(next(r['text'] for r in row['raw'] if r['path']=='report.json'))==row['report'];oracle.update(row['report']['counts'])
   if any(p['entries'] for p in row['processes']):executedUnits.append({'unit':row['unit'],'counts':{key:sum(p['counts'][key] for p in row['processes']) for key in ids},'totalHelperEntries':sum(p['entries'] for p in row['processes'])})
   for process in row['processes']:
    assert set(process['counts'])==set(ids);assert process['counts']==process['attempts'],'wrapped call and actual function entry disagree';assert process['entries']==sum(process['counts'].values())+sum(r['count'] for r in process['unknown'].values());counts.update(process['counts']);attempts.update(process['attempts']);total+=process['entries'];processes+=1
    for key,event in process['unknown'].items():
     if key not in unknown:unknown[key]={**event,'count':0}
     unknown[key]['count']+=event['count']
  units+=piece['units']
assert units==summary['units']==[u['name'] for u in plan['units']];assert len(units)==len(set(units))==398
assert dict(counts)==summary['counts'];assert dict(attempts)==summary['attempts'];assert total==summary['totalEntries'];assert processes==summary['nodeProcesses'];assert unknown==summary['unknownEntries'];assert dict(oracle)==summary['oracle']==plan['counts']
result={'complete':True,'acceptedUnitsExactlyOnce':398,'countsPerSourceCallSite':dict(counts),'wrappedCallsPerSourceSite':dict(attempts),'totalHelperEntries':total,'unknownEntries':unknown,'observedNodeProcesses':processes,'oracle':dict(oracle),'unitsWithEntries':executedUnits,'pieces':len(summary['pieces']),'sumPieceSeconds':sum(p['seconds'] for p in summary['pieces']),'largestPieceSeconds':max(p['seconds'] for p in summary['pieces']),'allArchiveAndInnerHashesVerified':True,'allRawGuardReportsMatch':True,'allCallAttemptsMatchActualEntries':True}
(s/'raw/runtime-reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
