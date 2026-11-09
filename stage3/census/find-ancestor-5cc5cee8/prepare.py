import subprocess,sys,os,json,re,hashlib
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];scratch=Path('/workspace/scratch');adapted=scratch/'bar-adapted'
os.environ['ADAMIC_TYPESCRIPT_SOURCE']=str(scratch/'bar-typescript')
def run(name,cwd,cmd,env=None):
 p=subprocess.run([sys.executable,str(out/'run.py'),name,str(cwd),*cmd],env=env);assert p.returncode==0,(name,p.returncode)
assert subprocess.check_output(['git','rev-parse','5cc5cee8^'],cwd=repo)==subprocess.check_output(['git','rev-parse','c1e9379e'],cwd=repo)
changed=subprocess.check_output(['git','diff','--name-only','c1e9379e','5cc5cee8'],cwd=repo,text=True).splitlines()
assert not any(p.startswith('stage3/adapt/') for p in changed)
(out/'pins.json').write_text(json.dumps({'base':subprocess.check_output(['git','rev-parse','c1e9379e'],text=True).strip(),'head':subprocess.check_output(['git','rev-parse','5cc5cee8'],text=True).strip(),'source':'050880ce59e30b356b686bd3144efe24f875ebc8','changed':changed},indent=2)+'\n')
run('apply',repo,['bash','stage3/apply.sh',str(adapted)])
(out/'patch-set.md').write_bytes((adapted/'patch-set.md').read_bytes())
source_hashes={str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((adapted/'src').rglob('*')) if p.is_file()}
(out/'source-hashes.json').write_text(json.dumps(source_hashes,indent=2)+'\n')
provenance={}
for side,pin in [('base','c1e9379e'),('head','5cc5cee8')]:
 tree=repo if side=='base' else scratch/'bar-head'
 if side=='head':
  run('head-worktree',repo,['git','worktree','add','--detach',str(tree),pin])
  (tree/'cohere').rmdir();(tree/'cohere').symlink_to(repo/'cohere',target_is_directory=True)
 work=scratch/('bar-'+side+'.work')
 pairs=re.findall(r'^\s*(\S+) => (\./cohere\S*)',(tree/'go.mod').read_text(),re.M)
 replacements={m:str(repo/p) for m,p in pairs if m!='github.com/microsoft/TypeScript/tsc'}
 work.write_text('go 1.27\n\ntoolchain go1.27.1\n\nuse (\n'+str(tree)+'\n'+str(repo/'cohere/TypeScript/tsc')+'\n)\n\nreplace (\n'+''.join(m+' => '+p+'\n' for m,p in replacements.items())+')\n')
 env=dict(os.environ,GOWORK=str(work));overlay=scratch/('bar-'+side+'-overlay');entry=scratch/('bar-'+side+'-entry')
 run(side+'-overlay',tree,['python3','stage3/census/latent/make_overlay.py',str(tree),str(overlay)],env)
 run(side+'-entry-overlay',tree,['python3','stage3/meter/entry_overlay.py',str(overlay),str(entry)],env)
 binary=scratch/('bar-'+side+'-bin')
 run(side+'-build',tree,['go','build','-p=2','-buildvcs=false','-overlay='+str(entry/'overlay.json'),'-o',str(binary),'./stage3/census/latent/tool'],env)
 run(side+'-overlay-tests',tree,['go','test','-count=1','-overlay='+str(overlay/'overlay.json'),'./stage3/census/latent/tool','-run','^TestLatent(Census|Stream|RelationOverflowHooks)'],env)
 provenance[side]={'sha':subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip(),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'checker':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo/'cohere',text=True).strip()}
 for label,directory in [('latent',overlay),('entry',entry)]:
  dest=out/'overlays'/side/label;dest.mkdir(parents=True);manifest=[]
  for i,(logical,actual) in enumerate(sorted(json.loads((directory/'overlay.json').read_text())['Replace'].items())):
   name=f'{i:02d}-'+Path(actual).name
   if name.endswith('.go'):name+='.txt'
   data=Path(actual).read_bytes();(dest/name).write_bytes(data);manifest.append(dict(logical=logical,evidence=name,sha256=hashlib.sha256(data).hexdigest()))
  (dest/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
print('prepare child_exit=0',flush=True)
