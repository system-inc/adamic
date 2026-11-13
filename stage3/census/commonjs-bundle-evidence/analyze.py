import collections,hashlib,json,subprocess
from pathlib import Path
out=Path('/workspace/scratch/995spev');prefix=str(out/'adapted/src/compiler')+'/'
def read(name):return [json.loads(line) for line in (out/name).read_text().splitlines()]
def norm(x):
 if isinstance(x,str):return x.replace(prefix,'')
 if isinstance(x,list):return [norm(y) for y in x]
 if isinstance(x,dict):return {k:norm(v) for k,v in x.items()}
 return x
rootrows={label:[norm(read(f'{label}-root-{i:02}.jsonl')[0]) for i in range(1,80)] for label in ['main','head']}
def stop(row):return tuple(row.get(k,'') for k in ['kind','where','reason','fix','diagnostics'])
moved=[dict(root=a['roots'][0],before=a,after=b) for a,b in zip(rootrows['main'],rootrows['head']) if stop(a)!=stop(b)]
latent={label:norm(read(label+'-sys.jsonl')) for label in ['main','head']}
def sites(rows):return {tuple(f.get(k,'') for k in ['kind','where','reason','text']) for row in rows for f in row.get('findings',[])}
def siteobj(f):return dict(zip(['kind','where','reason','text'],f))
a,b=sites(latent['main']),sites(latent['head'])
unit_findings={label:[f for row in rows for f in row.get('findings',[]) if f.get('unit')=='sys.ts:1475:5'] for label,rows in latent.items()}
units={label:[u for row in rows for u in row.get('units',[]) if u.get('name')=='getNodeSystem' or u.get('where','').startswith('sys.ts:1475:')] for label,rows in latent.items()}
reads={label:[f for f in sorted(sites(rows)) if f[1].startswith('sys.ts:1494:')] for label,rows in latent.items()}
result=dict(roots=79,moved=moved,sys_gone=list(map(siteobj,sorted(a-b))),sys_new=list(map(siteobj,sorted(b-a))),sys_totals={k:len(sites(v)) for k,v in latent.items()},getNodeSystem=units,getNodeSystem_findings=unit_findings,reads1494={k:list(map(siteobj,v)) for k,v in reads.items()})
(out/'comparison.json').write_text(json.dumps(result,indent=2)+'\n')
(out/'next-stops.json').write_text(json.dumps(rootrows,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='getNodeSystem'},indent=2))
