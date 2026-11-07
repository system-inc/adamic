from pathlib import Path
import os,subprocess,time,json
s=Path('/workspace/scratch/gov-b');phases={}
if not (s/'before').exists():
 with (s/'raw/prepare-copy.log').open('w') as log:subprocess.run(['node',s/'prepare.cjs'],stdout=log,stderr=log,check=True)
for side in ['before','after']:
 for name,args in [('build',['npm','run','build']),('harness',['npm','exec','hereby','--','tests'])]:
  begin=time.monotonic()
  with (s/'raw'/f'{side}-{name}.log').open('w') as log:r=subprocess.run(args,cwd=s/side,stdout=log,stderr=log,timeout=1400)
  phases[side+'-'+name]={'exit':r.returncode,'seconds':time.monotonic()-begin};(s/'raw/prepare.json').write_text(json.dumps(phases,indent=2)+'\n');assert r.returncode==0
 (s/side/'node_modules').unlink()
print('prepared',flush=True)
