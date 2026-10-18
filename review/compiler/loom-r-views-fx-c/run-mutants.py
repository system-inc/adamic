from pathlib import Path
import subprocess,json,difflib,time,os
root=Path.cwd(); ev=root/'review/compiler/loom-r-views-fx-c';results=[]
for row in json.loads((ev/'mutant-plan.json').read_text()):
 originals={};patches=[]
 try:
  for relative,before,after in row['changes']:
   p=root/relative;originals[p]=p.read_text();assert before in originals[p],(row['name'],relative)
   changed=originals[p].replace(before,after,1);p.write_text(changed)
   patches.extend(difflib.unified_diff(originals[p].splitlines(True),changed.splitlines(True),fromfile=relative,tofile=relative,n=0))
  (ev/(row['name']+'.patch')).write_text(''.join(patches).replace('+\t\n', '+\n'))
  command=['go','test',row['package'],'-run',row['test'],'-count=1','-v','-timeout','90s','-parallel','4']
  start=time.monotonic()
  with (ev/(row['name']+'.log')).open('w') as output:
   result=subprocess.run(command,stdout=output,stderr=subprocess.STDOUT,timeout=115,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'))
  text=(ev/(row['name']+'.log')).read_text()
  caught=result.returncode==0 if row['green'] else result.returncode!=0 and '--- FAIL:' in text and 'build failed' not in text and 'test timed out' not in text
  results.append(dict(name=row['name'],command=command,exit=result.returncode,expected_green=row['green'],verified=caught,seconds=round(time.monotonic()-start,3)))
  print(row['name'],caught,flush=True)
  assert caught,(row['name'],text[-3000:])
 finally:
  for p,original in originals.items():p.write_text(original)
  (ev/'mutant-results.json').write_text(json.dumps(results,indent=2)+'\n')
