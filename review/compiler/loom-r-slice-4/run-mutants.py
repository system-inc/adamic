import json,subprocess,time
from pathlib import Path
out=Path('review/compiler/loom-r-slice-4');rows=[]
commands=[['python3',str(out/'generic-mutants.py'),str(out/'generic-mutants')]]+[['python3',str(out/(name+'.py'))]+([str(out/'iteration-mutants')] if name=='iteration-mutants' else []) for name in ['generator-mutants','loader-mutants','iteration-mutants','extra-mutants','reconciliation-mutants','scout-tests']]
for i,cmd in enumerate(commands):
 log=out/('mutant-run-%d.log'%i);start=time.monotonic()
 with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=360)
 rows.append({'command':cmd,'exit':r.returncode,'seconds':time.monotonic()-start,'log':str(log)});(out/'mutant-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(i,r.returncode,rows[-1]['seconds'],flush=True)
 if r.returncode:raise SystemExit(1)
