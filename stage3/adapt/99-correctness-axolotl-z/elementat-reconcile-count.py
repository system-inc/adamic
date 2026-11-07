import hashlib,json,math,shutil
from collections import Counter,defaultdict
from pathlib import Path
s=Path('/workspace/scratch/elementat-b');out=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-z');summary=json.loads((out/'ELEMENTAT-COUNT.json').read_text());assert summary['complete']
counts=Counter();callers=defaultdict(Counter);distribution=Counter();offsets=set();nonintegers=0;files=0
for piece in summary['pieces']:
 directory=s/('piece-'+str(piece['piece']).zfill(2)+piece.get('attemptSuffix',''))/'calls'
 for file in sorted(directory.glob('*.jsonl')):
  assert file.name.rsplit('.',2)[0] in piece['units'];files+=1
  for line in file.open():
   row=json.loads(line);offset=row['offset'];length=row['length'];decoded={'NaN':float('nan'),'+Infinity':float('inf'),'-Infinity':float('-inf'),'-0':-0.0}.get(offset,offset) if isinstance(offset,str) else offset
   integer=isinstance(decoded,(int,float)) and math.isfinite(decoded) and float(decoded).is_integer()
   expected='array undefined' if length is None else 'non-integer' if not integer else 'out-of-range integer' if decoded>=length or decoded< -length else 'in-range negative integer' if decoded<0 else 'in-range non-negative integer'
   assert row['classification']==expected and row['nonIntegerOffset']==(not integer)
   counts[expected]+=1;distribution[(expected,str(offset),length)]+=1;offsets.add(str(offset));nonintegers+=not integer
   if expected!='in-range non-negative integer':
    assert row['caller'].startswith('src/'),row
    callers[expected][row['caller']]+=1
assert dict(counts)=={k:v for k,v in summary['counts'].items() if v}
assert nonintegers==summary['nonIntegerOffsetsIncludingUndefinedArrays']==0
for key,sites in summary['callers'].items():assert dict(callers[key])==sites
result={'complete':True,'calls':sum(counts.values()),'callFiles':files,'counts':summary['counts'],'callers':summary['callers'],'observedOffsets':sorted(offsets),'nonIntegerOffsetsIncludingUndefinedArrays':nonintegers,'independentClassificationAgrees':True,'exactlyOneCompleteUnitUnion':len(summary['completedUnits'])==len(set(summary['completedUnits']))==398,'counterVersionsSameObservationProtocol':True,'excludedAttempts':summary.get('excludedAttempts',[]),'excludedProbeDirectories':['seeds','probe-v2-calls','v2-identity-calls'],'distribution':[{'classification':c,'offset':o,'length':l,'count':n} for (c,o,l),n in sorted(distribution.items(),key=lambda x:(x[0][0],x[0][1],-1 if x[0][2] is None else x[0][2]))]}
(s/'raw/reconciled-count.json').write_text(json.dumps(result,indent=2)+'\n')
folder=s/'raw/caller-source-maps';folder.mkdir(exist_ok=True)
for p in (s/'counted/built/local').glob('*.js.map'):shutil.copyfile(p,folder/p.name)
assert (s/'count-preload.cjs').read_bytes()==(s/'count-preload-v2.cjs').read_bytes()
(s/'raw/recorder-versions.json').write_text(json.dumps({name:hashlib.sha256((s/name).read_bytes()).hexdigest() for name in ['count-preload-v1.cjs','count-preload-v2.cjs','count-preload.cjs','instrument.cjs','count-pieces.py']},indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k in ['calls','counts','observedOffsets','independentClassificationAgrees','exactlyOneCompleteUnitUnion']}))
