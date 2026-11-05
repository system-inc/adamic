import json,collections,hashlib,gzip
from pathlib import Path
root=Path('/workspace/scratch/99r')
identities=[]
for label in ['reference','counted']:
 report=json.loads((root/(label+'-oracle/report.json')).read_text())
 assert report['counts']=={'passing':106366,'failing':1,'pending':0},report['counts']
 assert report['baseline_diffs']==['api/typescript.d.ts'],report['baseline_diffs']
 assert report['workers']==8 and report['runners']=='all' and report['tests'] is None
 rows=collections.Counter((r['event'],r['title']) for line in (root/(label+'-identities.jsonl')).open() if (r:=json.loads(line)))
 assert all(n%2==0 for n in rows.values())
 assert sum(n for (event,title),n in rows.items() if event=='pass')==212732
 assert sum(n for (event,title),n in rows.items() if event=='fail')==2
 identities.append(rows)
assert identities[0]==identities[1],{'added':list((identities[1]-identities[0]).items()),'removed':list((identities[0]-identities[1]).items())}
normalized=''.join(json.dumps({'event':event,'title':title,'count':n//2},sort_keys=True)+'\n' for (event,title),n in sorted(identities[0].items()))
(root/'normalized-identities.jsonl').write_text(normalized)
counts=collections.defaultdict(collections.Counter);plants=collections.Counter();undefined=collections.defaultdict(collections.Counter)
for line in (root/'counts.jsonl').open():
 r=json.loads(line)
 if r['event']=='summary':
  for site,row in r['counts'].items(): counts[site].update(row)
 elif r['event']=='planted': plants[r['site']]+=1
 elif r['event']=='undefined': undefined[r['site']][r['test']]+=1
for site in ['14118','16980','41574','41628']:
 assert plants[site]==1
 assert counts['planted:'+site]=={'runs':1,'undefined':1}
 assert counts[site]['runs']>0
 assert sum(undefined[site].values())==counts[site]['undefined']
assert counts['14118']['undefined']>100
out=dict(counts=dict(counts),plants=dict(plants),undefinedTitles={s:dict(c) for s,c in undefined.items()},identityMatch=True,identitiesSha256=hashlib.sha256(normalized.encode()).hexdigest(),passing=106366,failing=1,pending=0)
(root/'count-result.json').write_text(json.dumps(out,indent=2)+'\n')
print(json.dumps({k:v for k,v in out.items() if k!='undefinedTitles'},indent=2))
