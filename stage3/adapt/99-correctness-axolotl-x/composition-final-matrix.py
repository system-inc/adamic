#!/usr/bin/env python3
"""Fresh full applies for the final call counter; preserve oracle applicability by byte identity."""
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import time
p=Path(__file__).resolve().parent;repo=p.parents[2];scratch=Path('/workspace/scratch/bind2');out=p/'composition';matrix=out/'matrix'
def wait(path):
 while not path.exists():time.sleep(5)
 return json.loads(path.read_text())
def run(args,cwd=repo):subprocess.run([str(x) for x in args],cwd=cwd,check=True)
def publish(message,*paths):run(['python3',p/'publish_paths.py',message,*[str(f.relative_to(repo)) for f in paths]])
assert wait(scratch/'oracle-done.json')['status']=='pass'
initial=scratch/'initial-matrix';shutil.copytree(matrix,initial)
archive=out/'initial-matrix.tar.gz';run(['python3',p/'composition-pack.py',initial,archive])
initial_reports={label:json.loads((matrix/(label+'.json')).read_text()) for label in ['x','32x','95hx','both']}
for label in initial_reports:
 for suffix in ['.json','.tar.gz','.sha256','.tar.gz.sha256']:
  f=matrix/(label+suffix)
  if f.exists():f.unlink()
 old=scratch/('apply-'+label);shutil.move(old,scratch/('initial-apply-'+label))
 tree=Path('/workspace/bind2-'+label+'-tree');shutil.rmtree(tree)
# Runtime adapter files are the only changed inputs in the fresh frozen snapshots.
roots={label:Path('/workspace/bind2-'+label+'-repo') for label in ['x','32x']};roots['95hx']=roots['both']=Path('/workspace/bind2-main322')
def update(root):
 for name in ['adapt.cjs','engine.cjs']:shutil.copyfile(p/name,root/'stage3/adapt/99-correctness-axolotl-x'/name)
 run(['git','add','stage3/adapt/99-correctness-axolotl-x/adapt.cjs','stage3/adapt/99-correctness-axolotl-x/engine.cjs'],root)
 run(['git','commit','-q','-m','Scratch final complete core-call counter'],root)
for root in {roots['x'],roots['32x'],roots['95hx']}:update(root)
# Reuse the standard matrix runner against fresh apply receipts.
with (scratch/'final-matrix-controller.log').open('w') as log:
 runner=subprocess.Popen(['python3',str(p/'composition-matrix.py'),'--95h'],cwd=repo,stdout=log,stderr=subprocess.STDOUT)
 try:
  def apply(label):run(['python3',p/'measure_piece.py',scratch/('apply-'+label),'--','bash',roots[label]/'stage3/apply.sh','--build-product','/workspace/bind2-'+label+'-tree'])
  with ThreadPoolExecutor(max_workers=2) as pool:
   futures=[pool.submit(apply,label) for label in ['x','32x']]
   for f in futures:f.result()
  target=roots['95hx']
  run(['git','restore','--source=322f8056','--staged','--worktree','stage3/adapt/32-indexed-reads-program'],target)
  run(['git','commit','-q','-m','Scratch final 95h plus 99x matrix'],target)
  apply('95hx')
  run(['git','restore','--source=997331ef1ec86480f47f3710fd43e1150407dafc','--staged','--worktree','stage3/adapt/32-indexed-reads-program'],target)
  run(['git','commit','-q','-m','Scratch final both-fold matrix'],target)
  apply('both')
  assert runner.wait()==0
 finally:
  if runner.poll() is None:runner.terminate();runner.wait()
# Both oracle artifacts must contain the exact final source and compiler products.
def digest(root):return {str(f.relative_to(root)):hashlib.sha256(f.read_bytes()).hexdigest() for f in root.rglob('*') if f.is_file()}
identity={}
for label in ['x','32x']:
 ready=list((scratch/('oracle-cache-'+label)).glob('*/ready.json'));assert len(ready)==1;artifact=ready[0].parent;tree=Path('/workspace/bind2-'+label+'-tree')
 assert digest(artifact/'tree/src')==digest(tree/'src')
 js={}
 for f in sorted((tree/'built/local').glob('*.js')):
  expected=artifact/'tree/built/local'/f.name;assert expected.read_bytes()==f.read_bytes();js[f.name]=hashlib.sha256(f.read_bytes()).hexdigest()
 assert len(js)==9
 identity[label]=dict(artifact=artifact.name,source_files_identical=len(digest(tree/'src')),all_nine_javascript_products_byte_identical=True,javascript_sha256=js,final_apply_commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=roots[label],text=True).strip())
for label,old in initial_reports.items():
 current=wait(matrix/(label+'.json'));assert current['status']=='pass'
 assert current['source']['expectedSites']==old['source']['expectedSites']
 for a,b in zip(old['javascript']['files'],current['javascript']['files']):assert a==b,(label,a['file'])
identityfile=out/'oracle-final-source-identity.json';identityfile.write_text(json.dumps(dict(status='pass',reason='Final validation counts parenthesized and other symbol-resolved core call expressions independently of rewrite eligibility.',matrix_unchanged=True,oracle_outputs_identical=identity,functional_sha256={name:hashlib.sha256((p/name).read_bytes()).hexdigest() for name in ['adapt.cjs','engine.cjs','owners.json','inventory.json','ownership.json']}),indent=2)+'\n')
publish('Final call-counter matrix passes; oracle source and all nine products are byte-identical',identityfile,archive,archive.with_suffix('').with_suffix('.sha256'),Path(str(archive)+'.sha256'))
(scratch/'final-matrix-done.json').write_text(json.dumps(dict(status='pass'))+'\n')
