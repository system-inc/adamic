import pathlib,hashlib,json,gzip,shutil
s=pathlib.Path('/workspace/scratch/path-mint-618');r={}
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(s/'base/built/local'/name).read_bytes();p=s/'lane/adapted-tree/built/local'/name;b=p.read_bytes()
 r[name]={'byteIdentical':a==b,'baseSha256':hashlib.sha256(a).hexdigest(),'afterSha256':hashlib.sha256(b).hexdigest(),'bytes':len(b)}
 with p.open('rb') as x,gzip.open(s/'evidence'/('after-'+name+'.gz'),'wb') as y:shutil.copyfileobj(x,y)
(s/'evidence/declarations-comparison.json').write_text(json.dumps(r,indent=2)+'\n');print(r)
assert all(x['byteIdentical'] for x in r.values())
