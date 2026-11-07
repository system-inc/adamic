import json,sys,hashlib
from collections import Counter,defaultdict
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');out=s/'full-count';sys.path.insert(0,str(repo/'stage3/lane'));from shard_plan import names_digest,load_plan
assert json.loads((out/'report.json').read_text())['status']=='pass';assert json.loads((s/'whole-control/report.json').read_text())['status']=='pass'
def records(p):return [dict(record,task=row['task'],status=status) for row in map(json.loads,p.read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
a,b=records(out/'tasks.jsonl'),records(s/'whole-control-tasks.jsonl');assert names_digest(a)==names_digest(b);assert (out/'oracle/baseline.diff').read_bytes()==(s/'whole-control/oracle/baseline.diff').read_bytes();plan=load_plan();counts=Counter(x['status'] for x in a);assert counts==Counter(plan['counts'])
calls=Counter();shapes={};kinds=defaultdict(Counter);outcomes=defaultdict(Counter)
for p in sorted((out/'calls').glob('*.jsonl')):
 for line in p.open():
  r=json.loads(line);owner=r['owner'];calls[owner]+=1;outcomes[owner][r['outcome']]+=1;assert r['comparisons']=={'typeOnlyIdenticalJavaScript':True}
  for i,shape in enumerate(r['shapes']):
   key=owner+':'+str(i+1);v=shapes.setdefault(key,{'plain':0,'notPlain':0,'inheritedEnumerable':0,'accessors':0,'symbolKeys':0});v['plain' if shape['plain'] else 'notPlain']+=1;kinds[key][shape['kind']]+=1
   for k,field in [('inheritedEnumerable','inheritedEnumerable'),('accessors','accessors'),('symbolKeys','symbols')]:v[k]+=bool(shape[field])
result={'complete':True,'scope':'One bounded, unchanged whole-lane oracle run, normal guards retained','counts':{k:counts[k] for k in plan['counts']},'calls':dict(calls),'shapes':shapes,'kinds':{k:dict(v) for k,v in kinds.items()},'outcomes':{k:dict(v) for k,v in outcomes.items()},'namesDigest':names_digest(a),'wholeLaneIdentitiesMatch':True,'wholeLaneRawBaselineDiffMatches':True,'counterexamples':0,'equivalenceProof':'Both selected owners, complete core JavaScript, all nine emitted JavaScript files, and both public declarations are byte-identical. Shape observation is not an alternative-implementation evaluation.','observerSha256':hashlib.sha256((s/'preload.cjs').read_bytes()).hexdigest()}
(s/'count-reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
