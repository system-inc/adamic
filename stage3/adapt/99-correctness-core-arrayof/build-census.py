from pathlib import Path
import json,subprocess
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');out=s/'overlay'
with (s/'overlay.log').open('w') as log:subprocess.run(['python3',str(repo/'stage3/census/latent/make_overlay.py'),str(repo),str(out)],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
p=out/'overlay.json';data=json.loads(p.read_text())
for k,v in data['Replace'].items():
 f=Path(v)
 if f.suffix=='.go':new=f.with_suffix('.go.txt');f.rename(new);data['Replace'][k]=str(new)
p.write_text(json.dumps(data,indent=2)+'\n')
with (s/'census-build.log').open('w') as log:subprocess.run(['go','build','-buildvcs=false','-overlay='+str(p),'-o',str(s/'latent'),'./stage3/census/latent/tool'],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
