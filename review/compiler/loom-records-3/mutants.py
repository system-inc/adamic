import subprocess,time,json,os
from pathlib import Path
out=Path(__file__).resolve().parent
root=out.parents[2]
env=os.environ.copy();env['RECORDS_REBUILD_MUTANT_LOGS']=str(out/'source-mutants');env['NAMED_INDEX_MUTANT_LOGS']=str(out/'named-mutants')
jobs=[('source-mutants',['python3','review/compiler/loom-records-2/run-source-mutants.py'],480),('named-mutants',['python3','review/compiler/loom-records-2/run-named-mutants.py'],240)]
rows=[]
for name,args,limit in jobs:
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:
  try:code=subprocess.run(args,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired:code=124
 rows.append(dict(name=name,command=args,exit=code,seconds=round(time.monotonic()-start,3)))
 (out/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(rows[-1]),flush=True)
 if code:break
