import json,os,shutil,signal,subprocess,time
from pathlib import Path
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/elementat-b/step-b');raw=s/'raw';tree=s/'after';phases={'apply':{'exit':0,'tree':str(tree)}}
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='',NODE_DISABLE_COMPILE_CACHE='1');env.pop('EA_CALL_DIR',None)
for name,args in [('install',['npm','ci','--no-audit','--no-fund']),('build',['npm','run','build'])]:
 begin=time.monotonic()
 with (raw/(name+'.log')).open('w') as log:
  child=subprocess.Popen(args,cwd=tree,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=child.wait(timeout=1400)
  finally:
   try:os.killpg(child.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   child.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-begin};(raw/'producer.json').write_text(json.dumps(phases,indent=2)+'\n');assert code==0,(name,code)
with (raw/'comparison.json').open('w') as log:
 run=subprocess.run(['node',str(s/'compare.cjs'),'/workspace/scratch/elementat-b/whole-control/adapted-tree',str(tree),str(raw)],env=env,stdout=log,stderr=subprocess.PIPE,text=True)
(raw/'comparison.log').write_text(run.stderr);assert run.returncode==0,run.stderr
packages=s/'packages';packages.mkdir();shutil.move(str(tree/'node_modules'),str(packages/'node_modules'))
(repo/'stage3/typescript/node_modules').symlink_to(packages/'node_modules',target_is_directory=True)
print('After build and strict initializer-only comparison pass',flush=True)
