import json,sys,hashlib
from collections import Counter,defaultdict
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import load_plan,names_digest
plan=load_plan();summary=json.loads((s/'count-summary.json').read_text());assert summary['complete'];assert summary['completedUnits']==[u['name'] for u in plan['units']];assert summary['oracleCounts']==plan['counts']
observed=[];rawdiff=b'';calls=Counter();shapes={};kinds=defaultdict(Counter);report_count=0;outcomes=defaultdict(Counter)
for index in range(1,21):
 piece=s/f'piece-{index:02d}'
 for unit in plan['units'][(index-1)*20:index*20]:
  folder=piece/unit['name'];r=json.loads((folder/'report.json').read_text());rows=json.loads((folder/'observations.json').read_text());assert r['status']=='pass' and r['counts']==unit['expect'];assert names_digest(rows)==unit['names_digest'];observed+=rows;rawdiff+=(folder/'oracle/baseline.diff').read_bytes();report_count+=1
 for file in sorted((piece/'calls').glob('*.jsonl')):
  for line in file.open():
   row=json.loads(line);owner=row['owner'];calls[owner]+=1;outcomes[owner][row['outcome']]+=1;assert row['comparisons']=={'typeOnlyIdenticalJavaScript':True}
   for i,shape in enumerate(row['shapes']):
    key=owner+':'+str(i+1);count=shapes.setdefault(key,{'plain':0,'notPlain':0,'inheritedEnumerable':0,'accessors':0,'symbolKeys':0});count['plain' if shape['plain'] else 'notPlain']+=1;kinds[key][shape['kind']]+=1
    for field,src in [('inheritedEnumerable','inheritedEnumerable'),('accessors','accessors'),('symbolKeys','symbols')]:count[field]+=bool(shape[src])
assert dict(calls)==summary['calls'];assert shapes==summary['shapes'];assert summary['comparisonFailures']==summary['uncompared']==0
whole=[dict(record,task=row['task'],status=status) for row in map(json.loads,(s/'whole-control-tasks.jsonl').read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
assert names_digest(observed)==names_digest(whole);assert rawdiff==(s/'whole-control/oracle/baseline.diff').read_bytes();assert json.loads((s/'whole-control/report.json').read_text())['status']=='pass'
(s/'count-reconciled.json').write_text(json.dumps({'complete':True,'units':report_count,'counts':summary['oracleCounts'],'calls':dict(calls),'shapes':shapes,'kinds':{k:dict(v) for k,v in kinds.items()},'outcomes':{k:dict(v) for k,v in outcomes.items()},'namesDigest':names_digest(observed),'wholeLaneIdentitiesMatch':True,'wholeLaneRawBaselineDiffMatches':True,'counterexamples':0,'equivalenceProof':'The selected type-only owners and all nine emitted JavaScript files are byte-identical. The observer records shapes only; independent Node witnesses also agree.','observerSha256':summary['observerSha256']},indent=2)+'\n')
print('reconciled',report_count,dict(calls),flush=True)
