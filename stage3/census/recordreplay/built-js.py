from pathlib import Path
import subprocess as s,os,json,hashlib,gzip
D=Path('/workspace/scratch/recordreplay');R=D/'base';O=R/'stage3/census/recordreplay/gate'
for label in ['control','candidate2']:
 for task in ['local']:
  s.run(['python3',str(D/'run.py'),label+'-built-'+task,'node',str(D/'control/node_modules/hereby/bin/hereby.js'),task],cwd=D/label,check=True)
products={}
for name in ['tsc.js','typescript.js','tsserverlibrary.js']:
 a=(D/'control/built/local'/name).read_bytes();b=(D/'candidate2/built/local'/name).read_bytes();assert a==b,name
 products[name]={'bytes':len(b),'sha256':hashlib.sha256(b).hexdigest(),'byte_identical':True}
for label in ['control','candidate2']:
 for task in ['local']:
  n=label+'-built-'+task
  (O/(n+'.json')).write_bytes((D/(n+'.json')).read_bytes());(O/(n+'.log.gz')).write_bytes(gzip.compress((D/(n+'.log')).read_bytes(),mtime=0))
(O/'built-javascript.json').write_text(json.dumps(products,indent=2)+'\n');print(products,flush=True)
