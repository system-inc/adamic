from pathlib import Path
import json,re,collections
r=Path('/workspace/scratch/write-file-slot-aprime')
def norm(s):
 s=re.sub(r'^.*?\.(?:ts|a):\d+:\d+:\s*','',s)
 s=s.replace('/workspace/scratch/write-file-slot-measure/main/','<tree>/').replace(str(r/'candidate')+'/','<tree>/')
 return re.sub(r'(?<=\.ts):\d+:\d+',':<pos>',s)
def read(p):
 rows=list(map(json.loads,p.read_text().splitlines()));assert rows[-1]['status']=='relation_report';h=rows[0];assert h['latent_mode']=='full' and h['module_format']=='CommonJSBundle' and not h['diagnostics'] and not h['checker_rejected']
 file=next(x for x in rows if 'units'in x);f=file['findings'];sites={tuple(x[k] for k in ['kind','phase','where','reason','text']):x for x in f if x['kind'] in ['Refused','NotYet']}
 groups=collections.defaultdict(list)
 for x in sites.values():groups[(x['kind'],x['phase'],x['reason'],norm(x['text']))].append(x)
 counts=collections.Counter((x['kind'],x['phase'],x['reason'],norm(x['text'])) for x in f)
 return {'header':h,'units':len(file['units']),'rawRecords':len(f),'uniqueSites':len(sites),'allNormalizedTuples':len(counts)},groups,counts,file
main,m,mc,mf=read(Path('/workspace/scratch/write-file-slot-measure/main.jsonl'));candidate,c,cc,cf=read(r/'full-sys.jsonl');gone=[];added=[]
for k in m.keys()|c.keys():
 a,b=m.get(k,[]),c.get(k,[])
 if len(a)>len(b):gone.extend(a[len(b):])
 if len(b)>len(a):added.extend(b[len(a):])
units={u['where']:u.get('name','') for u in cf['units']}
result={'main':main,'candidate':candidate,'newUnique':added,'goneUnique':gone,'changedUnique':0 if not added else 'see unpaired additions and removals','allRecordAdded':[{'key':list(k),'count':v} for k,v in (cc-mc).items()],'allRecordGone':[{'key':list(k),'count':v} for k,v in (mc-cc).items()],'candidateOwnerStops':[dict(f,unitName=units.get(f['unit'],'')) for f in cf['findings'] if units.get(f['unit'])=='getNodeSystem'],'pairKey':['kind','phase','reason','message without position prefix']}
(r/'pairs.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k!='candidateOwnerStops'},indent=2))
