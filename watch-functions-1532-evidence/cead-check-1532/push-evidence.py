from pathlib import Path
import os,subprocess,json,gzip,shutil,sys
s=Path('/tmp/watch-file-check-cead');stage=s/'evidence-work';out=stage/'watch-functions-1532-evidence/cead-check-1532';out.mkdir(parents=True,exist_ok=True)
for name in ['ownership-grep.txt','amend.cjs','compare-census.py','delta-after.json','delta-summary.log','run-census.py','gate-and-apply.py','count-run.py','retry-count.py','post-count.py','dynamic-control.cjs','scope-stock.cjs','witness.cjs','amend.log','api-install.log']:
 p=s/name
 if p.exists():shutil.copyfile(p,out/name)
for side in ['main','after']:
 (out/('sys-'+side+'.ts.gz')).write_bytes(gzip.compress((s/('sys-'+side+'.ts')).read_bytes(),mtime=0));dest=out/('census-'+side);dest.mkdir(exist_ok=True)
 for p in (s/('census-'+side)).iterdir():
  if p.suffix in ['.jsonl','.log']:(dest/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
  else:shutil.copyfile(p,dest/p.name)
for p in (s/'count-oracle-oom').glob('*'):
 if p.is_file():dest=out/'count-oracle-oom';dest.mkdir(exist_ok=True);(dest/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for name in ['apply-after.log','gate-and-apply.log','retry-count-controller.log','post-count.log','memory-events-before-count.txt','memory-events-after-count.txt']:
 p=s/name
 if p.exists():(out/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
patch=subprocess.check_output(['git','-C',str(s/'work'),'diff','ceadba1d9baf55dc0917eaf27dec69ebada795db']);(out/'implementation.patch.gz').write_bytes(gzip.compress(patch,mtime=0))
(out/'checkpoint.json').write_text(json.dumps({'base':'ceadba1d9baf55dc0917eaf27dec69ebada795db','census':{'new':0,'gone':1,'changed':0,'target1532Gone':True,'sibling1533Unchanged':True,'oldNodeWatchFileStopAbsent':True},'applyExit':(s/'apply-after-exit.txt').read_text() if (s/'apply-after-exit.txt').exists() else None,'count':'first run invalidated by OOM; quiet retry pending or running','status':'checkpoint; not a final proof verdict','noLoom':True},indent=2)+'\n')
subprocess.run([sys.executable,str(s/'archive-evidence.py')],check=True)
repo='/workspace/adamic';ref='refs/heads/typescript/watch-functions-1532-evidence';git=lambda args,**kw:subprocess.check_output(['git','-C',repo,*args],**kw).decode().strip()
parent=git(['rev-parse',ref]);index=s/'evidence-index';index.unlink(missing_ok=True);env=dict(os.environ,GIT_INDEX_FILE=str(index),GIT_WORK_TREE=str(stage))
for args in [['read-tree',parent],['add','--','watch-functions-1532-evidence/cead-check-1532']]:subprocess.run(['git','-C',repo,*args],env=env,check=True)
tree=git(['write-tree'],env=env);commit=git(['commit-tree',tree,'-p',parent,'-m',sys.argv[1] if len(sys.argv)>1 else 'Checkpoint ceadba1d watch-file interval census and proof work'],env=env)
subprocess.run(['git','-C',repo,'update-ref',ref,commit,parent],check=True);subprocess.run(['git','-C',repo,'push','origin',ref+':'+ref],check=True);print(commit)
