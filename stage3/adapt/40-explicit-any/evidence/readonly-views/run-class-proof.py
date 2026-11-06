import os,subprocess,json
from pathlib import Path
import sys
repo=Path(__file__).resolve().parents[5];source=Path(sys.argv[1]).resolve();work=Path(sys.argv[2]).resolve();work.mkdir(parents=True,exist_ok=True);tree=work/'class-tree';evidence=work/'class-proof';evidence.mkdir(exist_ok=True)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
def run(name,args):
 with (work/(name+'.log')).open('w') as log:
  p=subprocess.run(args,env=env,stdout=log,stderr=subprocess.STDOUT,cwd=repo)
 print(name,p.returncode,flush=True)
 if p.returncode:raise SystemExit(p.returncode)
if not tree.exists():run('class-clone',['git','clone','--no-checkout',str(source),str(tree)])
ref=subprocess.check_output(['git','-C',str(source),'rev-parse','refs/stage3/40-explicit-any-after'],text=True).strip()
run('class-archive',['git','-C',str(source),'archive','--format=tar','--output='+str(work/'class-input.tar'),ref])
run('class-checkout',['tar','-xf',str(work/'class-input.tar'),'-C',str(tree)])
(tree/'node_modules').symlink_to(source/'node_modules',target_is_directory=True)
rules=json.loads((repo/'stage3/adapt/40-explicit-any/class-rules.json').read_text())['enum-display']
for file in sorted({r['file'] for r in rules}):
 p=tree/file;s=p.read_bytes()
 for r in rules:
  if r['file']!=file:continue
  before=r['originalLine'];adapted=before[:r['column']-1]+r['replacement']+before[r['column']+2:]
  assert s.count(adapted.encode())==1,r['id']
  s=s.replace(adapted.encode(),before.encode(),1)
 p.write_bytes(s)
run('class-before-build',['npm','run','build:compiler','--prefix',str(tree)])
proof=str(repo/'stage3/adapt/40-explicit-any/class-proof.cjs')
run('class-before',['node',proof,'before',str(tree),str(evidence),'enum-display'])
run('class-adapt',['node',str(repo/'stage3/adapt/40-explicit-any/adapt.cjs'),str(tree)])
run('class-after-build',['npm','run','build:compiler','--prefix',str(tree)])
run('class-after',['node',proof,'after',str(tree),str(evidence),'enum-display'])
