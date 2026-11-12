import sys,os,subprocess,json,shutil,hashlib,time
from pathlib import Path
repo=Path('/workspace/path-components-adapt');base=Path('/workspace/scratch/path-adapt-after');tree=Path('/workspace/scratch/path-adapt-replay');out=Path('/workspace/scratch/path-adapt-proof')
shutil.copytree(base,tree,symlinks=True,ignore=lambda d,n: [x for x in ['node_modules','built'] if Path(d)==base and x in n])
(tree/'node_modules').symlink_to('/workspace/path-components-adapt/stage3/typescript/node_modules',target_is_directory=True)
def snapshot():
 result={}
 for folder in ['src','scripts','tests']:
  for f in (tree/folder).rglob('*'):
   if f.is_file():result[str(f.relative_to(tree))]=hashlib.sha256(f.read_bytes()).hexdigest()
 return result
initial=snapshot();rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
 start=time.monotonic()
 r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],env={**os.environ,'NODE_PATH':str(repo/'stage3/api/node_modules')},capture_output=True,text=True,timeout=600)
 current=snapshot();changes=[f for f in set(initial)|set(current) if initial.get(f)!=current.get(f)]
 row={'adapter':directory.name,'child_exit_code':r.returncode,'seconds':time.monotonic()-start,'stdout':r.stdout,'stderr':r.stderr,'changedFiles':changes};rows.append(row)
 (out/'replay.json').write_text(json.dumps(rows,indent=2)+'\n');print(directory.name+' child_exit='+str(r.returncode)+' changed='+str(len(changes)),flush=True)
 if r.returncode or changes:sys.exit(r.returncode or 1)
print('all adapters replayed; source/scripts/tests byte-identical',flush=True)
