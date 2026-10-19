from pathlib import Path
import json,subprocess,time,hashlib
out=Path(__file__).resolve().parent;root=out.parents[2];original=(root/'internal/lower/testdata/generic_body_relations/refused.a').read_bytes();header,_,body=original.partition(b'\n');scratch=Path('/workspace/scratch/slice4-header-mutants');scratch.mkdir(exist_ok=True);rows=[]
for name,data in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately absent reason\n'+body)]:
 path=scratch/(name+'.a');path.write_bytes(data);start=time.monotonic();command=['python3','-B',str(out/'check-headers.py'),name,str(path)]
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=180)
 result=json.loads((out/(name+'.json')).read_text());assert r.returncode==1 and list(result['paths'].values())[0]['status']=='refused' and result['failure']
 rows.append(dict(mutant=name,command=command,exit=r.returncode,catcher='unchanged Gate.aCheck',seconds=round(time.monotonic()-start,3),program_body_sha256=hashlib.sha256(body).hexdigest()))
(out/'header-mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
