import sys,os,shutil,json,hashlib,difflib
from pathlib import Path
sys.path.insert(0,'/workspace/scratch/elementat-declaration')
from run import run
s=Path('/workspace/scratch/elementat-declaration');baseline=s/'baseline';candidate=s/'candidate'
for p in baseline.iterdir():
 if p.name=='src':continue
 target=candidate/p.name
 if p.is_dir():shutil.copytree(p,target,symlinks=True)
 else:shutil.copy2(p,target)
(candidate/'node_modules').symlink_to(s/'packages/node_modules',target_is_directory=True)
r=run('candidate-compiler-build',['npm','run','build:compiler'],cwd=candidate);assert r['exit']==0,r
results={}
for name in ['typescript.d.ts','typescript.internal.d.ts','tsserverlibrary.d.ts']:
 a=(s/'counted/built/local'/name).read_bytes();b=(candidate/'built/local'/name).read_bytes();diff=''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='const/'+name,tofile='declaration/'+name))
 results[name]={'identical':a==b,'constSha256':hashlib.sha256(a).hexdigest(),'declarationSha256':hashlib.sha256(b).hexdigest(),'diff':diff}
(s/'declaration-emit.json').write_text(json.dumps({'compilerBuildExit':r['exit'],'files':results},indent=2)+'\n')
print(json.dumps({name:row['identical'] for name,row in results.items()}),flush=True)
