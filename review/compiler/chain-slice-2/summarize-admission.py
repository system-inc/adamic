import json,re
from pathlib import Path
p=Path('review/compiler/chain-slice-2')
inputs=json.loads((p/'admission-inputs.json').read_text())
observations={}
for side in ['main','slice']:
 rows={}
 for log in sorted(p.glob('admission-'+side+'-*.log')):
  if log.name in ['admission-main-2.log','admission-main-3.log'] and (p/log.name.replace('.log','-retry.log')).exists(): continue
  for line in log.read_text().splitlines():
   if not line.startswith('{'):continue
   event=json.loads(line)
   message=event.get('Output','')
   if 'ADMISSION ' not in message:continue
   path,outcome=message.split('ADMISSION ',1)[1].rstrip('\n').split('\t',1)
   assert path not in rows,path
   rows[path]=outcome
 assert set(rows)==set(inputs),(side,len(rows),len(inputs),sorted(set(inputs)-set(rows)))
 observations[side]=rows
changes=[{'path':path,'main':observations['main'][path],'slice':observations['slice'][path]} for path in inputs if observations['main'][path]!=observations['slice'][path]]
(p/'admission-delta.json').write_text(json.dumps({'inputs':len(inputs),'complete':True,'observations':observations,'changes':changes},indent=2)+'\n')
print(json.dumps({'inputs':len(inputs),'changes':len(changes)},indent=2))
for row in changes:print(json.dumps(row))
