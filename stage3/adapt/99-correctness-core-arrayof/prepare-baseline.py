from pathlib import Path
import subprocess,os,json,time,shutil,sys
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');tree=s/'before';phases={'apply':{'exit':0,'log':str(s/'apply-before.log'),'reusedCompletedFullApply':True}}
assert (tree/'patch-set.md').exists()
for name,command in [('install',['npm','ci','--no-audit','--no-fund']),('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
 start=time.monotonic();print('starting',name,flush=True)
 with (s/(name+'-before.log')).open('w') as log:r=subprocess.run(command,cwd=tree,stdout=log,stderr=subprocess.STDOUT,timeout=1700)
 phases[name]={'exit':r.returncode,'seconds':time.monotonic()-start};(s/'before-phases.json').write_text(json.dumps(phases,indent=2)+'\n');assert r.returncode==0
sys.path.insert(0,str(repo/'stage3/lane'));from artifact import input_key,payload_digest
artifact=s/'baseline-artifact';artifact.mkdir();(artifact/'tree').symlink_to(tree,target_is_directory=True);shutil.copytree(repo/'stage3/api/node_modules/typescript',artifact/'api');key,inputs=input_key();ready=dict(key=key,inputs=inputs,payload=payload_digest(artifact/'tree'),phases=phases,reuse='completed full baseline apply, then unchanged standard install/build/harness commands');(artifact/'ready.json').write_text(json.dumps(ready,indent=2)+'\n');print('READY',flush=True)
