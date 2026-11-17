import json,collections,pathlib,sys
kind,left,right,prefixLeft,prefixRight,out=sys.argv[1:]
def norm(v,prefix):
 if isinstance(v,str):return v.replace(prefix+'/', '')
 if isinstance(v,list):return [norm(x,prefix) for x in v]
 if isinstance(v,dict):return {k:norm(x,prefix) for k,x in v.items()}
 return v
def read(file,prefix):
 if kind=='scan':
  x=norm(json.load(open(file)),prefix)
  assert 'Scope' in x and not x.get('load_error'),x
  return x.get('Stops') or [],x
 if kind=='latent':
  rows=[norm(json.loads(s),prefix) for s in open(file) if s.strip()]
  assert rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle',rows[0]
  assert rows[-1]['status']=='relation_report',rows[-1]
  files=[r for r in rows if 'file' in r and 'findings' in r]
  assert len(files)==1,len(files)
  return files[0]['findings'],{'header':rows[0],'file':files[0]['file'],'units':files[0]['units'],'trailer':rows[-1]}
 raise ValueError(kind)
a,metaA=read(left,prefixLeft);b,metaB=read(right,prefixRight)
def key(x):return json.dumps({k:v for k,v in x.items() if k not in ['start','end']},sort_keys=True)
ca,cb=map(collections.Counter,([key(x) for x in a],[key(x) for x in b]))
gone=[json.loads(s) for s,n in (ca-cb).items() for _ in range(n)]
new=[json.loads(s) for s,n in (cb-ca).items() for _ in range(n)]
def location(x):return (x.get('Site',x.get('where')),x.get('Kind',x.get('kind')))
changed=[]
for loc in set(map(location,gone))&set(map(location,new)):
 changed.append({'site':loc,'before':[x for x in gone if location(x)==loc],'after':[x for x in new if location(x)==loc]})
newOnly=[x for x in new if location(x) not in set(map(location,gone))]
goneOnly=[x for x in gone if location(x) not in set(map(location,new))]
result={'kind':kind,'beforeEvents':len(a),'afterEvents':len(b),'new':newOnly,'gone':goneOnly,'changed':changed,'beforeMeta':metaA,'afterMeta':metaB}
pathlib.Path(out).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({'kind':kind,'before':len(a),'after':len(b),'new':len(newOnly),'gone':len(goneOnly),'changed':len(changed)}))
sys.exit(1 if newOnly or changed else 0)
