from pathlib import Path
import shutil,subprocess,json,os,time,hashlib
s=Path('/workspace/scratch/gov-a');repo=Path('/workspace/adamic');tree=s/'counted';started=time.monotonic()
assert subprocess.run(['git','diff','--quiet','8f4efea05','origin/main','--','stage3/adapt','stage3/api','stage3/apply.py','stage3/source.json'],cwd=repo).returncode==0
shutil.copytree('/workspace/scratch/gov/before',tree,symlinks=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
with (s/'raw/apply-99z.log').open('w') as log:subprocess.run(['node',repo/'stage3/adapt/99-correctness-axolotl-z/adapt.cjs',tree],cwd=repo,env=env,stdout=log,stderr=log,check=True)
(s/'raw/base.json').write_text(json.dumps({'main':subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip(),'prerequisite':'efa9c7dd7ffd4504068e6beeae26a2365eb145f1','mergedBase':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),'sourceReuse':'complete earlier full apply; origin/main adaptation inputs byte-identical to 8f4efea05, then exact 99z applied','coreBeforeCounterSha256':hashlib.sha256((tree/'src/compiler/core.ts').read_bytes()).hexdigest()},indent=2)+'\n')
with (s/'raw/instrument.json').open('w') as log:subprocess.run(['node',s/'instrument.cjs',tree],env=env,stdout=log,stderr=log,check=True)
phases={}
for name,args in [('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
 begin=time.monotonic()
 with (s/'raw'/f'{name}.log').open('w') as log:r=subprocess.run(args,cwd=tree,stdout=log,stderr=log,timeout=1400)
 phases[name]={'exit':r.returncode,'seconds':time.monotonic()-begin};(s/'raw/prepare.json').write_text(json.dumps(phases,indent=2)+'\n');assert r.returncode==0
print('prepared seconds',time.monotonic()-started,flush=True)
