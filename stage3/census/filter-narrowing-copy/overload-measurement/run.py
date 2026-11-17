#!/usr/bin/env python3
"""Report-only exact-owner controls. Arguments: baseline tree, scratch, scan, latent."""
import os,sys,json,subprocess,hashlib
from pathlib import Path
baseline,out,scan,latent=map(lambda x:Path(x).resolve(),sys.argv[1:])
out.mkdir(parents=True,exist_ok=True)
source=(baseline/'src/compiler/core.ts').read_text()
lines=source.splitlines()
start=next(i for i,x in enumerate(lines) if x.startswith('export function filter<'))
implementation=next(i for i,x in enumerate(lines) if x.startswith('export function filter<') and x.endswith(' {'))
end=next(i for i in range(implementation+1,len(lines)) if lines[i]=='}')
signatures=[(i,x) for i,x in enumerate(lines[start:implementation],start) if x.startswith('export function filter<')]
assert len(signatures)==8
cases=[(str(k),[k]) for k in range(1,9)]+[('boolean',[2,4,6,8]),('readonly',[3,4,7,8])]
rows=[]
for name,keep in cases:
 directory=out/name;directory.mkdir(exist_ok=True)
 control=['']*(end+1)
 for k,(i,text) in enumerate(signatures,1):
  if k in keep:control[i]=text
 control[implementation:end+1]=lines[implementation:end+1]
 file=directory/'core.ts';file.write_text('\n'.join(control)+'\n')
 row={'case':name,'kept_original_overloads':keep,'source_sha256':hashlib.sha256(file.read_bytes()).hexdigest()}
 for tool,args in [('scan',[str(scan),str(file)]),('latent',[str(latent),str(directory),str(directory/'latent.jsonl')])]:
  env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(file),CENSUS_MODULE_FORMAT='CommonJSBundle')
  r=subprocess.run(args,env=env,capture_output=True,text=True)
  (directory/(tool+'.stdout')).write_text(r.stdout);(directory/(tool+'.stderr')).write_text(r.stderr);(directory/(tool+'.exit')).write_text(str(r.returncode)+'\n')
  row[tool+'_exit']=r.returncode
  if r.returncode: print(json.dumps(row),flush=True);sys.exit(r.returncode)
 observation=json.loads((directory/'scan.stdout').read_text());row['scan']=observation
 row['latent']=list(map(json.loads,(directory/'latent.jsonl').read_text().splitlines()))
 rows.append(row);print(name,json.dumps(observation.get('Stops')),flush=True)
(out/'results.json').write_text(json.dumps({'signatures':[{'overload':k,'line':i+1,'text':s} for k,(i,s) in enumerate(signatures,1)],'cases':rows},indent=2)+'\n')
