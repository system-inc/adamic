import json,re,collections
from pathlib import Path
r=Path('/workspace/scratch/write-file-slot-run')
def strip(s):
 return re.sub(r'(?:[^\s:]+[/\\])*[^\s:]+\.(?:ts|a):\d+:\d+:?\s*','',s)
def key(f):return tuple([f.get('kind'),f.get('phase'),f.get('reason'),strip(f.get('text',''))])
def read(label):
 rows=[json.loads(s) for s in (r/(label+'.jsonl')).read_text().splitlines()]
 assert rows[-1].get('status')=='relation_report'
 files=[x for x in rows if 'units' in x];assert len(files)==1
 assert not rows[0]['diagnostics'] and not rows[0]['checker_rejected']
 assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle'
 return rows,files[0]
base,b=read('main');result={}
for label in ['main','a','b','hidden']:
 rows,f=read(label);counts=collections.Counter(key(x) for x in f['findings']);old=collections.Counter(key(x) for x in b['findings'])
 units={x['where']:x.get('name',x.get('kind')) for x in f['units']}
 def row(k,n):return dict(zip(['kind','phase','reason','message'],k),count=n)
 added=[row(k,n) for k,n in (counts-old).items()];gone=[row(k,n) for k,n in (old-counts).items()]
 stops=[]
 for x in f['findings']:
  if x['phase']=='refusal_scan':
   stops.append(dict(x,unit_name=units.get(x['unit'])))
 result[label]={'header':rows[0],'units':len(f['units']),'findings':len(f['findings']),'unique':len(counts),'added':added,'gone':gone,'scan':stops,'all_findings':f['findings']}
(r/'summary.json').write_text(json.dumps(result,indent=2))
for label,x in result.items():
 print(label,x['findings'],x['unique'],'added',len(x['added']),'gone',len(x['gone']))
 print('ADDED',json.dumps(x['added'])); print('GONE',json.dumps(x['gone']))
 for s in x['scan']:
  if s['unit_name'] in ['patchWriteFileEnsuringDirectory','getNodeSystem'] or 'writeFile' in s['reason']:print('STOP',json.dumps(s))
