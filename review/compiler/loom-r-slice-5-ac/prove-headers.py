"""Exercise the actual pinned predicate and compiler with two first-line mutants."""
from pathlib import Path
import json,subprocess,sys,hashlib
out=Path(__file__).resolve().parent;root=out.parents[2];row=json.loads((out/'headers.json').read_text())[0];path=root/row['path'];original=path.read_bytes();body=original.split(b'\n',1)[1];results=[]
for label,contents in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately wrong slice5 header reason\n'+body)]:
 try:
  path.write_bytes(contents)
  command=[sys.executable,str(out/'check-headers.py'),label,row['path']]
  with (out/'logs'/(label+'.txt')).open('w') as log:r=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=180)
  measured=json.loads((out/(label+'.json')).read_text());assert r.returncode==1 and measured['exits']['a-check']==1 and measured['paths'][row['path']]['status']=='refused' and measured['failure']
  results.append(dict(mutant=label,path=row['path'],command=command,exit=r.returncode,caught_by='unchanged Gate.aCheck using actual branch compiler',failure=measured['failure'],body_sha256=hashlib.sha256(body).hexdigest()))
 finally:path.write_bytes(original)
(out/'header-mutants.json').write_text(json.dumps(results,indent=2)+'\n');print('PASS removed-header and wrong-reason mutants; original restored')
