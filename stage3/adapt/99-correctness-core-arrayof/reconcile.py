from pathlib import Path
import json,sys,collections,hashlib,math
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');sys.path.insert(0,str(repo/'stage3/lane'));from shard_plan import load_plan,names_digest
plan=load_plan();summary=json.loads((s/'COUNT.json').read_text());assert summary['complete'];assert summary['completedUnits']==[u['name'] for u in plan['units']]
counts=collections.Counter();histogram=collections.Counter();returned=0;bad=0;thrown=0;observations=[];totals=collections.Counter();diff=b''
for unit in plan['units']:
 d=s/'results'/unit['name'];report=json.loads((d/'report.json').read_text());assert report['status']=='pass' and report['counts']==unit['expect'];rows=json.loads((d/'observations.json').read_text());assert names_digest(rows)==unit['names_digest'];observations+=rows;totals.update(report['counts']);diff+=(d/'oracle/baseline.diff').read_bytes()
for piece in sorted(s.glob('piece-*')):
 for f in sorted((piece/'calls').glob('*.jsonl')):
  for line in f.open():
   row=json.loads(line)
   if row['event']=='call':
    value=row['count'];classification='NaN' if value=='NaN' else 'infinite' if value in ['Infinity','-Infinity'] else 'fractional' if isinstance(value,float) and not value.is_integer() else 'negative' if value<0 else 'non-negative integer'
    assert classification==row['classification'],row
    counts[classification]+=1;histogram[str(value)]+=1
   else:returned+=1;bad+=not row['equal'];thrown+='callbackThrewAt' in row
assert dict(counts)=={k:v for k,v in summary['counts'].items() if v};assert returned==summary['results']==sum(counts.values());assert bad==summary['counterexamples']==0;assert counts['non-negative integer']==returned>0
control=s/'control';assert json.loads((control/'verdict.json').read_text())['status']=='pass'
whole=[dict(record,task=row['task'],status=status) for row in map(json.loads,(control/'whole-tasks.jsonl').read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
assert names_digest(whole)==names_digest(observations);assert dict(totals)==plan['counts'];assert diff==(control/'oracle/baseline.diff').read_bytes()
result=dict(complete=True,units=len(plan['units']),counts=summary['counts'],returnedOrThrew=returned,callbackThrows=thrown,counterexamples=bad,oracleCounts=dict(totals),countHistogram=dict(histogram),independentlyReclassifiedEveryRawInput=True,testIdentityDigest=names_digest(whole),sameExactTestMultiset=True,baselineDiffByteIdentical=True,maxPieceSeconds=max(p['seconds'] for p in summary['pieces']))
(s/'RECONCILED.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
