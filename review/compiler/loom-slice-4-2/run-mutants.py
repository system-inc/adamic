import json,subprocess,os,time
from pathlib import Path
out=Path('review/compiler/loom-slice-4-2');rows=[]
cmds=[['python3','review/compiler/chain-slice-4/generic-mutants.py',str(out/'generic-mutants')]]+[['python3',str(out/(name+'.py'))]+([str(out/'iteration-mutants')] if name=='iteration-mutants' else []) for name in ['generator-mutants','loader-mutants','iteration-mutants','extra-mutants','supplement-mutants']]
for i,cmd in enumerate(cmds):
 start=time.monotonic();log=out/('mutant-run-'+str(i)+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=360)
 rows.append({'command':cmd,'exit':r.returncode,'seconds':time.monotonic()-start});print(i,r.returncode,rows[-1]['seconds'],flush=True)
 if r.returncode:break
# Four independent measurement controls and six deliberate assertion failures.
for i,old in enumerate(json.loads(Path('review/compiler/chain-slice-4/scout-measurement.json').read_text())):
 cmd=old['command'];log=out/('scout-'+str(i)+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=30)
 rows.append({'command':cmd,'exit':r.returncode,'expected':old['exit']});assert r.returncode==old['exit'],cmd
(out/'mutant-results.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(r['exit']==r.get('expected',0) for r in rows)
