from pathlib import Path
import json,difflib,re,hashlib
r=Path('/workspace/scratch/binding-views');out=r/'evidence'
def findings(name,file,mode):
 d=[json.loads(x) for x in (out/name/(file+'.jsonl')).read_text().splitlines()]
 meta=json.loads((out/name/(file+'.run.json')).read_text());assert meta['exit']==0
 assert d[0]['latent_mode']==mode and d[0]['module_format']=='CommonJSBundle';assert d[-1]['status']=='relation_report';assert d[-1]['module_format']=='CommonJSBundle'
 records=[row for row in d if row.get('file')];assert len(records)==1
 if mode=='scan only':assert not records[0]['units']
 result=records[0]['findings']+d[-1].get('findings',[])
 if mode=='scan only':assert all(row['phase']=='refusal_scan' for row in result)
 return result

def compare(name,file,mode):
 a=(r/'baseline/src/compiler'/file).read_text().splitlines();b=(r/name/'src/compiler'/file).read_text().splitlines();mapping={}
 for op,i,j,k,l in difflib.SequenceMatcher(None,a,b,autojunk=False).get_opcodes():
  if op=='equal' or (op=='replace' and j-i==l-k):mapping.update({line+1:i+(line-k)+1 for line in range(k,l)})
 def normalize(row,after):
  n=dict(row)
  def loc(m):
   line=int(m[1]);return 'src/compiler/'+file+':'+str(mapping.get(line,line) if after else line)+':'+m[2]
  for field in ['where','text','unit','reason']:
   s=n.get(field,'').replace(str(r/name)+'/', '').replace(str(r/'baseline')+'/', '')
   n[field]=re.sub(r'src/compiler/'+re.escape(file)+r':(\d+):(\d+)',loc,s)
  return n
 def index(rows,after):
  result={}
  for row in rows:
   n=normalize(row,after);result[tuple(n.get(k,'') for k in ['phase','kind','where','reason'])]=n
  return result
 old=findings('baseline',file,mode);new=findings(name,file,mode);aa=index(old,False);bb=index(new,True)
 result=dict(candidate=name,file=file,mode=mode,before_raw=len(old),after_raw=len(new),before_unique=len(aa),after_unique=len(bb),new=sorted([bb[k] for k in bb.keys()-aa.keys()],key=lambda x:x['where']),gone=sorted([aa[k] for k in aa.keys()-bb.keys()],key=lambda x:x['where']),changed=[dict(before=aa[k],after=bb[k]) for k in aa.keys()&bb.keys() if aa[k]['text']!=bb[k]['text']])
 return result,aa,bb
rows=[]
for file,mode,lines in [('resolutionCache.ts','full',[774,1053,1056,1471]),('program.ts','scan only',[2796,2829])]:
 result,base,candidate=compare('candidate',file,mode);assert not result['new'] and not result['changed'],result
 gone=result['gone'];assert len(gone)==len(lines),gone
 assert {int(x['where'].split(':')[-2]) for x in gone}==set(lines),gone
 assert all(x['phase']=='refusal_scan' and x['kind']=='Refused' for x in gone)
 mutant,_,restored=compare('mutant',file,mode);assert not mutant['new'] and not mutant['changed'] and not mutant['gone'],mutant
 assert all(key in restored for key in base.keys()-candidate.keys())
 assert (r/'baseline/src/compiler'/file).read_bytes()==(r/'mutant/src/compiler'/file).read_bytes()
 (out/(file+'-delta.json')).write_text(json.dumps(result,indent=2)+'\n');(out/(file+'-mutant.json')).write_text(json.dumps(mutant,indent=2)+'\n')
 rows.append(dict(file=file,mode=mode,before_raw=result['before_raw'],after_raw=result['after_raw'],before_unique=result['before_unique'],after_unique=result['after_unique'],new=0,changed=0,gone=len(gone),restored_targets=len(gone)))
(out/'census-summary.json').write_text(json.dumps(dict(status='pass',rows=rows),indent=2)+'\n');print(json.dumps(rows,indent=2))
