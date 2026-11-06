#!/usr/bin/env python3
"""Verify completed full applies, publish small summaries and packed raw receipts."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import time
p=Path(__file__).resolve().parent;repo=p.parents[2];scratch=Path('/workspace/scratch/bind2');out=p/'composition/matrix';out.mkdir(parents=True,exist_ok=True)
def wait(name):
 f=scratch/('apply-'+name)/'metrics.json'
 while not f.exists():time.sleep(5)
 value=json.loads(f.read_text());assert value['exit']==0,value
 return value
def run(name,command,cwd,raw,env=None):
 with (raw/(name+'.log')).open('w') as stream:r=subprocess.run(command,cwd=cwd,env=env,stdout=stream,stderr=subprocess.STDOUT)
 assert r.returncode==0,(name,r.returncode);return r.returncode
def publish(message,*files):subprocess.run(['python3',str(p/'publish_paths.py'),message,*[str(f.relative_to(repo)) for f in files]],cwd=repo,check=True)
for label,folds in [('main',''),('x',''),('32x','32')]+([('95hx','95h'),('both','32,95h')] if '--95h' in sys.argv else []):
 receipt=out/(label+'.json')
 if receipt.exists():continue
 metrics=wait(label);tree=Path('/workspace/bind2-'+('main-tree' if label=='main' else label+'-tree'));raw=scratch/('matrix-'+label);raw.mkdir(exist_ok=True)
 shutil.copyfile(scratch/('apply-'+label)/'process.log',raw/'apply.log');shutil.copyfile(scratch/('apply-'+label)/'metrics.json',raw/'apply-metrics.json')
 modules=Path('/workspace/bind-after/node_modules');assert (tree/'package-lock.json').read_bytes()==Path('/workspace/bind-after/package-lock.json').read_bytes()
 if not (tree/'node_modules').exists():(tree/'node_modules').symlink_to(modules,target_is_directory=True)
 roots=sorted(str(f) for f in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 begin=time.monotonic();run('types',[str(scratch/'bin/adamic'),'types',*roots],repo,raw);types=dict(exit=0,roots=len(roots),seconds=time.monotonic()-begin,command=[str(scratch/'bin/adamic'),'types',*roots],binary_sha256=hashlib.sha256((scratch/'bin/adamic').read_bytes()).hexdigest());(raw/'types.json').write_text(json.dumps(types,indent=2)+'\n')
 run('build',['npm','run','build:compiler'],tree,raw)
 result=dict(status='pass',label=label,folds=folds,full_apply=metrics,types=dict(exit=0,roots=79,diagnostics=0,seconds=types['seconds']),build_exit=0)
 if label!='main':
  env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
  run('source',['node',str(p/'composition-source.cjs'),'/workspace/bind2-main-tree',str(tree),str(raw),folds],repo,raw,env)
  run('javascript',['node',str(p/'composition-compare.cjs'),'/workspace/bind2-main-tree',str(tree),str(raw/'javascript'),folds],repo,raw,env)
  result['source']=json.loads((raw/'source-comparison.json').read_text());result['javascript']=json.loads((raw/'javascript/comparison.json').read_text());result['declarations']=json.loads((raw/'javascript/declarations.json').read_text())
  for name in (tree/'built/local').glob('*.js'):
   r=subprocess.run(['diff','-u','--label','main/'+name.name,'--label',label+'/'+name.name,str(Path('/workspace/bind2-main-tree/built/local')/name.name),str(name)],stdout=subprocess.PIPE);assert r.returncode in [0,1];(raw/(name.name+'.diff')).write_bytes(r.stdout)
 archive=out/(label+'.tar.gz');subprocess.run(['python3',str(p/'composition-pack.py'),str(raw),str(archive)],check=True)
 receipt.write_text(json.dumps(result,indent=2)+'\n')
 publish('99x composition matrix '+label+' passes full apply, types and emitted scope',receipt,archive,archive.with_suffix('').with_suffix('.sha256'),Path(str(archive)+'.sha256'))
