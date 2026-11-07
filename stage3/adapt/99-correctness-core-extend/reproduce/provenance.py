import hashlib,json,os,subprocess
from pathlib import Path
s=Path('/workspace/scratch/core-extend');repo=Path('/workspace/adamic')
def output(*args):return subprocess.check_output(args,cwd=repo,text=True).strip()
def h(p):return hashlib.sha256(p.read_bytes()).hexdigest()
result={'base':output('git','rev-parse','origin/main'),'sourcePin':output('git','-C','/workspace/typescript-source','rev-parse','HEAD'),'node':output('node','--version'),'go':output('go','version'),'binaries':{p.name:h(p) for p in [s/'adamic',s/'latent'] if p.exists()},'ruling':'Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts','class':'type-only','ownerNames':['extend','copyProperties'],'setup':'source /workspace/adamic-tools/env.sh; bash cloud/setup.sh; npm ci --prefix stage3/api','sourceEnvironment':'ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source'}
(s/'provenance.json').write_text(json.dumps(result,indent=2)+'\n')
