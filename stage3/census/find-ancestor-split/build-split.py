import os, subprocess, sys, json, re, hashlib
from pathlib import Path
out=Path(__file__).resolve().parent
checker=Path('/workspace/adamic/cohere')
def run(name,tree,cmd,env):
 rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(tree),*cmd],env=env).returncode
 print(name+' wrapper_exit='+str(rc),flush=True)
 if rc:sys.exit(rc)
provenance={}
for side,sha in [('a','7aabf373'),('b','52a75455')]:
 tree=Path('/workspace/scratch/fa-'+side)
 pin=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip()
 assert pin.startswith(sha)
 assert subprocess.run(['git','diff','--quiet',sha,'--','internal','stage3/census/latent','stage3/meter'],cwd=tree).returncode==0
 if not (tree/'cohere').is_symlink():
  (tree/'cohere').rmdir();(tree/'cohere').symlink_to(checker,target_is_directory=True)
 pairs=re.findall(r'^\s*(\S+) => (\./cohere\S*)',(tree/'go.mod').read_text(),re.M)
 replacements={m:str(tree/p) for m,p in pairs if m!='github.com/microsoft/TypeScript/tsc'}
 work=Path('/workspace/scratch/fa-'+side+'.work')
 work.write_text('go 1.27\n\ntoolchain go1.27.1\n\nuse (\n'+str(tree)+'\n'+str(checker/'TypeScript/tsc')+'\n)\n\nreplace (\n'+''.join(m+' => '+p+'\n' for m,p in replacements.items())+')\n')
 env=dict(os.environ,GOWORK=str(work));overlay=Path('/workspace/scratch/fa-'+side+'-overlay');entry=Path('/workspace/scratch/fa-'+side+'-entry')
 run(side+'-overlay',tree,['python3','stage3/census/latent/make_overlay.py',str(tree),str(overlay)],env)
 run(side+'-entry-overlay',tree,['python3','stage3/meter/entry_overlay.py',str(overlay),str(entry)],env)
 binary=Path('/workspace/scratch/fa-'+side+'-entry-bin')
 run(side+'-entry-build',tree,['go','build','-p=2','-buildvcs=false','-overlay='+str(entry/'overlay.json'),'-o',str(binary),'./stage3/census/latent/tool'],env)
 dest=out/'overlays'/side;dest.mkdir(parents=True,exist_ok=True);manifest=[]
 for i,(logical,actual) in enumerate(sorted(json.loads((entry/'overlay.json').read_text())['Replace'].items())):
  name=f'{i:02d}-'+Path(actual).name
  if name.endswith('.go'):name+='.txt'
  data=Path(actual).read_bytes();(dest/name).write_bytes(data)
  manifest.append(dict(logical=logical,evidence=name,sha256=hashlib.sha256(data).hexdigest()))
 (dest/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
 (out/(side+'-workspace.txt')).write_bytes(work.read_bytes())
 provenance[side]=dict(sha=pin,checker=subprocess.check_output(['git','rev-parse','HEAD'],cwd=checker,text=True).strip(),binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest())
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
print('builds complete child_exit=0',flush=True)
