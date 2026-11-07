"""Reopen every accepted piece and reconcile the frozen whole-oracle inventory."""
from pathlib import Path
import json,tarfile,hashlib,collections
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/debug-messages');e=s/'evidence';plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());summary=json.loads((e/'before-lane.json').read_text());assert summary['complete'];units=[];oracle=collections.Counter();counts=collections.Counter();processes=0
for piece in summary['pieces']:
 assert not piece['failures'] and piece['seconds']<1500;archive=e/piece['archive'];assert hashlib.sha256(archive.read_bytes()).hexdigest()==piece['sha256'];assert archive.stat().st_size<20_000_000
 with tarfile.open(archive) as tf:
  for line in tf.extractfile('SHA256SUMS').read().decode().splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(tf.extractfile(name).read()).hexdigest()==digest
  observed=json.loads(tf.extractfile('piece.json').read());assert all(observed[k]==piece[k] for k in observed)
  rows=[json.loads(l) for l in tf.extractfile('observations.jsonl')];assert sorted(r['unit'] for r in rows)==sorted(piece['units'])
  for row in rows:
   assert row['child_exit']==0 and row['report']['status']=='pass' and not row['report']['errors'];assert row['processes'];subtotal=collections.Counter()
   for process in row['processes']:assert set(process['counts'])=={'assertEachNode','assertNode','assertNotNode','assertOptionalNode'};subtotal.update(process['counts']);processes+=1
   assert dict(subtotal)==row['counts'] and not any(subtotal.values());counts.update(subtotal);oracle.update(row['report']['counts'])
   # Reconcile the raw ordinary guard report, rather than trusting only its copy.
   report=json.loads(next(r['text'] for r in row['raw'] if r['path']=='report.json'));assert report==row['report']
  units+=piece['units']
assert units==summary['units']==[u['name'] for u in plan['units']];assert len(units)==len(set(units))==398;assert dict(oracle)==summary['oracle']==plan['counts'];assert dict(counts)==summary['counts'] and not any(counts.values())
result={'wholeOracleCompleted':True,'acceptedUnitsExactlyOnce':398,'planSha256':hashlib.sha256((repo/'stage3/lane/shard-plan.json').read_bytes()).hexdigest(),'counts':dict(counts),'oracle':dict(oracle),'observedNodeProcesses':processes,'archivesVerified':len(summary['pieces']),'largestPieceSeconds':max(p['seconds'] for p in summary['pieces'])}
(s/'raw/whole-lane-reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
