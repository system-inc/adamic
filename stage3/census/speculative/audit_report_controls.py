"""Prove stock source/depth checks and checker-preservation checks reject real artifacts."""
from copy import deepcopy
import json, subprocess, sys
from pathlib import Path
root,raw,stock,full,no_stubs,out=map(lambda x:Path(x).resolve(),sys.argv[1:7])
out.mkdir(parents=True,exist_ok=True)
here=Path(__file__).resolve().parent
rows=[json.loads(x) for x in raw.read_text().splitlines()]
witness=json.loads(stock.read_text())
def execute(name,arguments,should_pass):
 with (out/(name+'.log')).open('w') as log:
  result=subprocess.run(['python3',*map(str,arguments)],stdout=log,stderr=subprocess.STDOUT)
 text=(out/(name+'.log')).read_text()
 if should_pass:assert result.returncode==0,text
 else:assert result.returncode!=0 and 'AssertionError' in text,text
 print(name+': '+('PASS' if should_pass else 'caught'))
execute('stock-baseline',[here/'report.py',root,raw,stock,out/'baseline'],True)
for name in ['depth-zero','stock-version','source-hash','checker-diagnostics','checker-spans']:
 changed=deepcopy(rows);proof=deepcopy(witness)
 if name=='depth-zero':
  for row in changed[1:]:
   for finding in row['findings']:finding['depth']=0
 elif name=='stock-version':proof['typescript']='6.0.4'
 elif name=='source-hash':proof['files'][0]['sha256']='mutant'
 elif name=='checker-diagnostics':changed[0]['diagnostics'].append('mutant checker diagnostic')
 else:changed[0]['diagnostic_sites'].append({'mutant':True})
 observed=out/(name+'.jsonl');evidence=out/(name+'.json')
 observed.write_text(''.join(json.dumps(row)+'\n' for row in changed))
 evidence.write_text(json.dumps(proof)+'\n')
 if name.startswith('checker-'):
  arguments=[here/'compare.py',full,no_stubs,observed]
 else:arguments=[here/'report.py',root,observed,evidence,out/(name+'-results')]
 execute(name+'-mutant',arguments,False)
