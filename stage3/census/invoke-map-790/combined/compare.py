import json,collections,pathlib,difflib,sys
s=pathlib.Path('/tmp/invoke-map-4pc8wd8');ev=s/'evidence-2'
afile=pathlib.Path(sys.argv[1]);bfile=pathlib.Path(sys.argv[2]);asrc=pathlib.Path(sys.argv[3]).read_bytes();bsrc=pathlib.Path(sys.argv[4]).read_bytes();out=pathlib.Path(sys.argv[5])
al,bl=asrc.splitlines(True),bsrc.splitlines(True)
ap=[0];bp=[0]
for line in al:ap.append(ap[-1]+len(line))
for line in bl:bp.append(bp[-1]+len(line))
ops=[(t,ap[i],ap[j],bp[k],bp[l]) for t,i,j,k,l in difflib.SequenceMatcher(None,al,bl,autojunk=False).get_opcodes()]
refined=[]
for t,i,j,k,l in ops:
 if t=='replace':
  refined.extend((u,i+a,i+b,k+c,k+d) for u,a,b,c,d in difflib.SequenceMatcher(None,asrc[i:j],bsrc[k:l],autojunk=False).get_opcodes())
 else:refined.append((t,i,j,k,l))
ops=refined
def oldoffset(n):
 for tag,i,j,k,l in ops:
  if k<=n<=l:
   if tag=='equal':return i+n-k
   if n==l:return j
   return i
 return n

def read(p):
 rs=[json.loads(l) for l in p.read_text().splitlines()];assert rs[0]['latent_mode']=='full' and rs[0]['module_format']=='CommonJSBundle';assert rs[-1]['status']=='relation_report';return rs

def fs(rs,after=False):
 out=[]
 for row in rs:
  for f in row.get('findings',[]):
   f=dict(f)
   if after and f.get('where','').rsplit(':',2)[0].endswith('/compiler/sys.ts'):
    for key in ('start','end'):
     if isinstance(f.get(key),int):f[key]=oldoffset(f[key])
   out.append(f)
 return out
ar,br=read(afile),read(bfile);a,b=fs(ar),fs(br,True)
ac=collections.Counter(json.dumps(f,sort_keys=True) for f in a);bc=collections.Counter(json.dumps(f,sort_keys=True) for f in b)
gone=[json.loads(k) for k,n in (ac-bc).items() for _ in range(n)];new=[json.loads(k) for k,n in (bc-ac).items() for _ in range(n)];changed=[]
def key(f):return tuple(f.get(k) for k in ('where','phase','unit','kind'))
for g in gone[:]:
 match=next((n for n in new if key(n)==key(g)),None)
 if match:changed.append({'main':g,'head':match});gone.remove(g);new.remove(match)
result={'main_observations':len(a),'head_observations':len(b),'new':new,'gone':gone,'changed':changed,'byte_offset_mapping':[{'tag':t,'main':[i,j],'head':[k,l]} for t,i,j,k,l in ops if t!='equal']}
out.write_text(json.dumps(result,indent=2)+'\n')
for label in ('new','gone','changed'):
 print(label+': '+str(len(result[label])))
 for f in result[label]:print(json.dumps(f,sort_keys=True))
