from pathlib import Path
import subprocess,os,json,time
s=Path('/workspace/scratch/gov-b');tree=s/'after';p=tree/'node_modules'
if not p.exists():p.symlink_to('/workspace/adamic/stage3/typescript/node_modules',target_is_directory=True)
env=dict(os.environ,NODE_OPTIONS='--require='+str(s/'preload.cjs'));data=json.loads((s/'raw/prepare.json').read_text())
for name,args in [('after-final-build',['npm','run','build']),('after-final-harness',['npm','exec','hereby','--','tests'])]:
 begin=time.monotonic()
 with (s/'raw'/(name+'-verified.log')).open('w') as log:r=subprocess.run(args,cwd=tree,env=env,stdout=log,stderr=log,timeout=1400)
 data[name]={'exit':r.returncode,'seconds':time.monotonic()-begin};(s/'raw/prepare.json').write_text(json.dumps(data,indent=2)+'\n');assert r.returncode==0
p.unlink()
