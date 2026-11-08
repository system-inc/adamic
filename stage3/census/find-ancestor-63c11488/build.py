import os,subprocess,sys,json,re,hashlib
from pathlib import Path
out=Path(__file__).resolve().parent
parent=Path('/workspace/scratch/fa-parent');head=Path('/workspace/scratch/fa-head')
# Shared pinned checker, with absolute replacements in separate workspaces.
if not (head/'cohere').is_symlink():
 (head/'cohere').rmdir();(head/'cohere').symlink_to(parent/'cohere',target_is_directory=True)
for side,tree in [('parent',parent),('head',head)]:
 work=Path('/workspace/scratch/fa-'+side+'.work')
 pairs=re.findall(r'^\s*(\S+) => (\./cohere\S*)',(tree/'go.mod').read_text(),re.M)
 replacements={m:str(parent/p) for m,p in pairs if m!='github.com/microsoft/TypeScript/tsc'}
 work.write_text('go 1.27\n\ntoolchain go1.27.1\n\nuse (\n'+str(tree)+'\n'+str(parent/'cohere/TypeScript/tsc')+'\n)\n\nreplace (\n'+''.join(m+' => '+p+'\n' for m,p in replacements.items())+')\n')
 env=dict(os.environ,GOWORK=str(work))
 overlay=Path('/workspace/scratch/fa-'+side+'-overlay');entry=Path('/workspace/scratch/fa-'+side+'-entry')
 def run(name,cmd):
  rc=subprocess.run([sys.executable,str(out/'run.py'),side+'-'+name,str(tree),*cmd],env=env).returncode
  if rc:raise SystemExit(rc)
 run('overlay',['python3','stage3/census/latent/make_overlay.py',str(tree),str(overlay)])
 run('entry-overlay',['python3','stage3/meter/entry_overlay.py',str(overlay),str(entry)])
 for name,directory,package in [('full',overlay,'./stage3/census/latent/tool'),('entry',entry,'./stage3/census/latent/tool')]:
  run(name+'-build',['go','build','-p=2','-buildvcs=false','-overlay='+str(directory/'overlay.json'),'-o','/workspace/scratch/fa-'+side+'-'+name,package])
 meta=json.loads((entry/'overlay.json').read_text())
 key=next(k for k in meta['Replace'] if k.endswith('/stage3/census/latent/replay/worker/main.go'))
 meta['Replace'][key]=str(out/'focus-driver.go.txt')
 focus=Path('/workspace/scratch/fa-'+side+'-focus-overlay.json');focus.write_text(json.dumps(meta,indent=2)+'\n')
 run('focus-build',['go','build','-p=2','-buildvcs=false','-overlay='+str(focus),'-o','/workspace/scratch/fa-'+side+'-focus','./stage3/census/latent/replay/worker'])
 for label,directory in [('full',overlay),('entry',entry)]:
  dest=out/'overlays'/side/label;dest.mkdir(parents=True,exist_ok=True);manifest=[]
  for i,(logical,actual) in enumerate(sorted(json.loads((directory/'overlay.json').read_text())['Replace'].items())):
   name=f'{i:02d}-'+Path(actual).name
   if name.endswith('.go'):name+='.txt'
   data=Path(actual).read_bytes();(dest/name).write_bytes(data)
   manifest.append(dict(logical=logical,evidence=name,sha256=hashlib.sha256(data).hexdigest()))
  (dest/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
 (out/(side+'-workspace.txt')).write_bytes(work.read_bytes())
print('builds complete child_exit=0',flush=True)
