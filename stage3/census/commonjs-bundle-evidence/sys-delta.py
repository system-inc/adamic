import collections,json
from pathlib import Path
out=Path('/workspace/scratch/995spev');prefix=str(out/'adapted/src/compiler')+'/'
def normalized(x):
 if isinstance(x,str):return x.replace(prefix,'')
 if isinstance(x,list):return list(map(normalized,x))
 if isinstance(x,dict):return {k:normalized(v) for k,v in x.items()}
 return x
def read(label):return normalized([json.loads(line) for line in (out/(label+'-sys.jsonl')).read_text().splitlines()])
def sites(rows):return {tuple(f.get(k,'') for k in ['kind','where','reason','text']) for row in rows for f in row.get('findings',[])}
def obj(f):return dict(zip(['kind','where','reason','text'],f))
rows={label:read(label) for label in ['main','head']}
for label,data in rows.items():
 assert data[0]['latent_mode']=='full'
 assert len(data)==3 and data[-1]['status']=='relation_report'
 assert data[1]['file']=='sys.ts'
a,b=sites(rows['main']),sites(rows['head'])
result=dict(gone=list(map(obj,sorted(a-b))),new=list(map(obj,sorted(b-a))),unique_by_kind={label:dict(collections.Counter(f[0] for f in sites(data))) for label,data in rows.items()},getNodeSystem={label:dict(units=[u for row in data for u in row.get('units',[]) if u.get('name')=='getNodeSystem'],findings=[f for row in data for f in row.get('findings',[]) if f.get('unit')=='sys.ts:1475:5']) for label,data in rows.items()})
(out/'sys-delta.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='getNodeSystem'},indent=2))
