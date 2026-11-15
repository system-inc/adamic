import json,re,collections
from pathlib import Path
r=Path('/workspace/scratch/write-file-slot-measure')
def strip(s):
 s=re.sub(r'^.*?\.(?:ts|a):\d+:\d+:\s*','',s)
 s=re.sub(r'/workspace/scratch/write-file-slot-measure/(?:main|a-construction|b-slot)/','<tree>/',s)
 return re.sub(r'(?<=\.ts):\d+:\d+',':<pos>',s)
def read(n):
 rows=[json.loads(l) for l in (r/(n+'.jsonl')).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle' and not rows[0]['diagnostics'] and not rows[0]['checker_rejected']
 assert rows[-1]['status']=='relation_report'
 records=[x for x in rows if 'file' in x and 'units' in x]
 raw=[f for x in records for f in x['findings'] if f['kind'] in ['Refused','NotYet']]
 unique={tuple(f[k] for k in ['kind','phase','where','reason','text']):f for f in raw}
 keys=collections.defaultdict(list)
 for f in unique.values():keys[tuple([f['kind'],f['phase'],f['reason'],strip(f['text'])])].append(f)
 units={u['where']:u for x in records for u in x['units']}
 owners={x['where']:x for x in records for x in x.get('declarations',[])}
 relevant=[]
 for f in raw:
  named=units.get(f['unit'],{}).get('name','')
  if named in ['patchWriteFileEnsuringDirectory','getNodeSystem'] or 'writeFile' in f['reason']: relevant.append(dict(f,unit_name=named))
 return {'header':rows[0],'raw_count':len(raw),'unique_site_phase_count':len(unique),'unique_site_count':len({(f['kind'],f['where'],f['reason'],f['text']) for f in raw}),'phase_counts':dict(collections.Counter(f['phase'] for f in unique.values())),'units':list(units.values()),'relevant':relevant},keys
main,m=read('main');out={'main':main}
for n in ['a-construction','b-slot']:
 data,c=read(n);added=[];gone=[]
 for k in m.keys()|c.keys():
  old=m.get(k,[]);new=c.get(k,[])
  if len(new)>len(old):added.extend(new[len(old):])
  if len(old)>len(new):gone.extend(old[len(new):])
 data['unpaired_added']=added;data['unpaired_gone']=gone
 data['pair_key']=['kind','phase','reason','text with positions and tree prefix stripped']
 out[n]=data
(r/'pairs.json').write_text(json.dumps(out,indent=2)+'\n')
print(json.dumps({n:{k:v for k,v in x.items() if k not in ['units','relevant']} for n,x in out.items()},indent=2))
