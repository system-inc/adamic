import hashlib,json,os,subprocess,sys,time
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/census/next-stops-b322b6ba'
sys.path.insert(0,str(out));from run import checkpoint
for name in ['apply','build']:
 p=out/(name+'-status.json')
 while not p.exists():time.sleep(1)
 assert json.loads(p.read_text())['exit_code']==0,name
binary=Path('/workspace/scratch/next-stops-census')
(out/'binaries.json').write_text(json.dumps({'ordinary':{'path':str(binary),'sha256':hashlib.sha256(binary.read_bytes()).hexdigest()}},indent=2)+'\n')
with (out/'census-buildinfo.txt').open('w') as f:
 result=subprocess.run(['go','version','-m',str(binary)],stdout=f,stderr=f)
 print('buildinfo child_exit='+str(result.returncode),flush=True);result.check_returncode()
tree=Path('/workspace/scratch/next-stops-adapted')
(out/'apply-source-manifest.json').write_text(json.dumps({str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()},indent=2)+'\n')
checkpoint('apply')
with (out/'groups-runner.log').open('w') as f:
 result=subprocess.run([sys.executable,str(out/'run.py'),'groups'],stdout=f,stderr=f)
print('groups child_exit='+str(result.returncode),flush=True)
(out/'groups-status.json').write_text(json.dumps({'exit_code':result.returncode},indent=2)+'\n')
sys.exit(result.returncode)
