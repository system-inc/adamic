import sys,os,time,json,shutil,hashlib
from pathlib import Path
sys.path.insert(0,'/workspace/scratch/elementat-declaration')
from run import run
s=Path('/workspace/scratch/elementat-declaration');repo=Path('/workspace/adamic');out=repo/'stage3/census/elementat-declaration'
def wait(name):
 while True:
  rows=[json.loads(l) for l in (s/'commands.jsonl').read_text().splitlines() if json.loads(l)['name']==name]
  if rows:assert rows[-1]['exit']==0,rows[-1];return
  time.sleep(1)
wait('apply')
def do(name,args,**kw):
 r=run(name,args,**kw);assert r['exit']==0,r
base=s/'baseline';candidate=s/'candidate';counted=s/'counted'
manifest={str(p.relative_to(base/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((base/'src').rglob('*')) if p.is_file()}
(s/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
shutil.copytree(base/'src',candidate/'src')
p=candidate/'src/compiler/core.ts';text=p.read_bytes().decode();pin=json.loads((repo/'stage3/adapt/99-correctness-axolotl-z/owners.json').read_text())[0]['after'];nl='\r\n' if '\r\n' in text else '\n';pin=pin.replace('\n',nl);assert text.count(pin)==1
new=(s/'declaration.txt').read_text().strip().replace('\n',nl)
p.write_bytes(text.replace(pin,new).encode());(s/'candidate-owner.txt').write_text(new+'\n')
changed=[str(p.relative_to(candidate/'src')) for p in (candidate/'src').rglob('*') if p.is_file() and hashlib.sha256(p.read_bytes()).hexdigest()!=manifest[str(p.relative_to(candidate/'src'))]];assert changed==['compiler/core.ts'],changed
(s/'source-delta.json').write_text(json.dumps({'changed':changed,'bodyUnchanged':True,'baseline':manifest['compiler/core.ts'],'candidate':hashlib.sha256(p.read_bytes()).hexdigest()})+'\n')
do('scan',['node',str(out/'scan.cjs'),str(base),str(s/'static.json')])
shutil.copytree(base,counted,symlinks=True)
do('instrument-reads',['node',str(out/'scan.cjs'),str(counted),str(s/'instrumented-reads.json'),'--instrument'])
do('instrument-body',['node',str(out/'instrument.cjs'),str(counted)])
do('lane-install',['npm','ci','--no-audit','--no-fund'],cwd=counted)
# Product shard views link this standard package destination.
env=dict(os.environ,NODE_OPTIONS='--require='+str(out/'preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1')
do('lane-build',['npm','run','build'],cwd=counted,env=env)
do('lane-harness',['npm','exec','hereby','--','tests'],cwd=counted,env=env)
(s/'packages').mkdir();shutil.move(str(counted/'node_modules'),str(s/'packages/node_modules'))
modules=repo/'stage3/typescript/node_modules';modules.parent.mkdir(exist_ok=True);assert not modules.exists();modules.symlink_to(s/'packages/node_modules',target_is_directory=True)
(s/'lane-ready.json').write_text(json.dumps({'ready':True,'tree':str(counted)})+'\n')
