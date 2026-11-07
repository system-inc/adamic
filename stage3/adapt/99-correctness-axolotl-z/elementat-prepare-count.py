import json,os,shutil,subprocess,time
from pathlib import Path
s=Path('/workspace/scratch/elementat-b');raw=s/'raw';tree=s/'counted';repo=Path('/workspace/adamic');started=time.monotonic();phases={}
# The prior complete applied source is byte-identical: main changed no stage3
# inputs relative to 7f40251e; a fresh full apply is independently in progress.
shutil.copytree('/workspace/scratch/elementat/before',tree,symlinks=True)
with (raw/'instrument.json').open('w') as log:subprocess.run(['node',str(s/'instrument.cjs'),str(tree)],stdout=log,stderr=subprocess.STDOUT,check=True)
env=dict(os.environ,NODE_OPTIONS='--require='+str(s/'count-preload.cjs')+' --enable-source-maps')
for name,args in [('install',['npm','ci','--no-audit','--no-fund']),('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
 begin=time.monotonic()
 with (raw/(name+'.log')).open('w') as log:
  result=subprocess.run(args,cwd=tree,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=1400,start_new_session=True)
 phases[name]={'exit':result.returncode,'seconds':time.monotonic()-begin};(raw/'prepare.json').write_text(json.dumps(phases,indent=2)+'\n')
 assert result.returncode==0,(name,result.returncode)
print('prepared',time.monotonic()-started,flush=True)
