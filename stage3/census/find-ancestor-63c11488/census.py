import os,subprocess,sys,json,hashlib,time
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];adapted=Path('/workspace/scratch/fa-adapted')
baseenv={k:v for k,v in os.environ.items() if not k.startswith('LATENT_')}
baseenv.update(LATENT_ASSERT_NO_OUTPUT='1',GOMEMLIMIT='5GiB',GOGC='70')
def checkpoint(event):
 rc=subprocess.run([sys.executable,str(out/'checkpoint.py'),event],cwd=repo).returncode
 if rc:raise SystemExit(rc)
def launch(side,name,kind,args,flags):
 env=dict(baseenv,**flags,LATENT_TIMES=str(out/(name+'-units.jsonl')))
 return subprocess.Popen([sys.executable,str(out/'run.py'),name,str(repo),'/workspace/scratch/fa-'+side+'-'+kind+('-bin' if kind=='entry' else ''),*args],env=env)
def finish(p,name):
 rc=p.wait();print(name,'wrapper_exit='+str(rc),flush=True)
 if rc:raise SystemExit(rc)
 checkpoint(name)
def hashes():
 return {str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((adapted/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
before=hashes();(out/'source-hashes.json').write_text(json.dumps(before,indent=2)+'\n')
provenance={}
for side,sha in [('parent','1060360b'),('head','63c11488')]:
 tree=Path('/workspace/scratch/fa-'+side)
 pin=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip()
 assert pin==subprocess.check_output(['git','rev-parse',sha],cwd=tree,text=True).strip()
 assert subprocess.run(['git','diff','--quiet',sha,'--','internal','stage3/census/latent','stage3/meter'],cwd=tree).returncode==0
 provenance[side]=dict(sha=pin,checker=subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree/'cohere',text=True).strip(),binaries={kind:hashlib.sha256(Path('/workspace/scratch/fa-'+side+'-'+kind+('-bin' if kind=='entry' else '')).read_bytes()).hexdigest() for kind in ['full','entry','focus']})
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
for file in ['utilitiesPublic.ts','core.ts']:
 for side in ['parent','head']:
  name=side+'-full-'+file.removesuffix('.ts')
  p=launch(side,name,'full',[str(adapted/'src/compiler'),str(out/(name+'.jsonl'))],dict(LATENT_FULL='1',LATENT_ONLY_FILE=str(adapted/'src/compiler'/file)))
  finish(p,name)
# The old report's full-unit replay selector, preserving the complete entry graph.
with (out/'sites.json').open('w') as f:
 rc=subprocess.run(['node',str(out/'sites.cjs'),str(adapted)],env=dict(baseenv,NODE_PATH=str(repo/'stage3/api/node_modules')),stdout=f).returncode
print('sites child_exit='+str(rc),flush=True);assert rc==0
sites=json.loads((out/'sites.json').read_text());units={r['unit']:r['site'] for r in reversed(sites['checker'])}
units.update({r['site']:r['site'] for r in sites['findAncestor'] if r['has_body']})
index={u:i for i,u in enumerate(sorted(units),1)};(out/'focus-index.json').write_text(json.dumps(index,indent=2)+'\n')
for side in ['parent','head']:
 for u,selector in sorted(units.items()):
  name=side+'-focus-'+str(index[u])
  p=launch(side,name,'focus',['-project',str(adapted/'src/tsc/tsc.ts'),'-where',selector,'-output',str(out/(name+'.jsonl'))],dict(LATENT_FULL='1'))
  finish(p,name)
# Simultaneous entry runs under a memory limit, identical settings on each side.
active={}
for side in ['parent','head']:
 name=side+'-entry-run'
 active[name]=launch(side,name,'entry',[str(adapted/'src/tsc/tsc.ts'),str(out/(name+'.jsonl'))],{})
last=0
while active:
 for name,p in list(active.items()):
  if p.poll() is not None:
   finish(p,name);del active[name]
 if time.monotonic()-last>60:
  print('entry active '+','.join(active),flush=True);last=time.monotonic()
 time.sleep(1)
assert hashes()==before,'adapted sources changed'
for label,prefix in [('entry','entry-run'),('utilitiesPublic','full-utilitiesPublic'),('core','full-core')]:
 cmd=['python3',str(repo/'stage3/census/pair/pair.py'),str(out/('parent-'+prefix+'.jsonl')),str(out/('head-'+prefix+'.jsonl')),str(adapted),'--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/(label+'-paired.json')),'--report',str(out/(label+'-paired.md'))]
 rc=subprocess.run([sys.executable,str(out/'run.py'),label+'-pair',str(repo),*cmd],env=baseenv).returncode
 if rc:raise SystemExit(rc)
print('all census runs and pairings complete child_exit=0',flush=True)
