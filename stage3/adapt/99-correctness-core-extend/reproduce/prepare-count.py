import os,time,subprocess,shutil,json,signal
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic');before=s/'before';tree=s/'counted'
while not (before/'patch-set.md').exists():
 print('waiting for complete baseline apply',flush=True);time.sleep(30)
while any('python3 ' in line and '/stage3/apply.py --build-product /workspace/scratch/core-extend/' in line and line.strip().split(None,1)[1].startswith('python3 ') for line in subprocess.check_output(['ps','-eo','pid,args'],text=True).splitlines()[1:]):
 print('waiting for full product applies to exit',flush=True);time.sleep(10)
with (s/'count-install.log').open('w') as log:subprocess.run(['npm','ci','--no-audit','--no-fund'],cwd=before,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=300)
shutil.copytree(before,tree,symlinks=True,ignore=lambda directory,names: [n for n in names if Path(directory)==before and n in ['.git','node_modules','built']])
(tree/'node_modules').symlink_to(before/'node_modules',target_is_directory=True)
(tree/'.git').symlink_to(before/'.git',target_is_directory=True)
env=dict(os.environ,NODE_OPTIONS='--require='+str(s/'preload.cjs'))
phases={}
for name,args in [('instrument',['node',str(s/'instrument.cjs'),str(tree)]),('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
 begin=time.monotonic()
 with (s/('count-'+name+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=tree,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=1400)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-begin};(s/'count-preparation.json').write_text(json.dumps(phases,indent=2)+'\n');print(name,phases[name],flush=True)
 if code:raise SystemExit(code)

# Shards supply their own pinned modules in each private view.
(tree/"node_modules").unlink()
