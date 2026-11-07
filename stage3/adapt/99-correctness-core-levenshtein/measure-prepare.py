import json,os,shutil,subprocess,time,signal
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');tree=s/'counted'
if not tree.exists():shutil.copytree(s/'before',tree,symlinks=True)
if (tree/'src').exists():shutil.rmtree(tree/'src')
shutil.copytree(s/'before/src',tree/'src')
if (tree/'built').exists():shutil.rmtree(tree/'built')
packages=s/'measure-packages';packages.mkdir(exist_ok=True)
if not (tree/'node_modules').exists() and (packages/'node_modules').exists():(tree/'node_modules').symlink_to(packages/'node_modules',target_is_directory=True)
needInstall=not (tree/'node_modules').exists()
env=dict(os.environ,NODE_OPTIONS='--require='+str(s/'preload.cjs'));phases={}
for name,args in [('instrument-v2',['node',str(s/'instrument.cjs'),str(tree)]),*([('measure-install',['npm','ci','--no-audit','--no-fund'])] if needInstall else []),('measure-build',['npm','run','build']),('measure-harness',['npm','exec','hereby','--','tests'])]:
 begin=time.monotonic()
 with (s/(name+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=tree,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=1700)
  except subprocess.TimeoutExpired:code=124
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-begin};(s/'measure-prepare.json').write_text(json.dumps(phases,indent=2)+'\n');print(name,phases[name],flush=True);assert code==0
if (tree/'node_modules').is_symlink():(tree/'node_modules').unlink()
else:
 assert not (packages/'node_modules').exists(),'external packages already exist';shutil.move(str(tree/'node_modules'),str(packages/'node_modules'))
modules=repo/'stage3/typescript/node_modules';modules.parent.mkdir(exist_ok=True)
if modules.is_symlink():modules.unlink()
assert not modules.exists(),'refuse to replace a package directory'
modules.symlink_to(packages/'node_modules',target_is_directory=True)
