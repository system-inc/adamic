import json,subprocess,time,os
from pathlib import Path
out=Path('review/compiler/loom-slice-4-2'); selections=json.loads((out/'test-selection.json').read_text()); results=[]
for pkg,names in selections.items():
 if pkg=='internal/ir': continue
 for i in range(0,len(names),12):
  group=names[i:i+12]; log=out/(pkg.replace('/','-')+'-'+str(i//12)+'.jsonl')
  cmd=['go','test','./'+pkg,'-run','^('+'|'.join(group)+')$','-parallel','4','-count=1','-timeout','90s','-json']; start=time.monotonic()
  with log.open('w') as f:
   try: code=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=180).returncode
   except subprocess.TimeoutExpired: code=124
  row={'command':cmd,'exit':code,'seconds':time.monotonic()-start,'log':str(log)}; results.append(row); print(pkg,i,code,row['seconds'],flush=True)
  if code: break
(out/'focused-results.json').write_text(json.dumps(results,indent=2)+'\n')
raise SystemExit(int(any(r['exit'] for r in results)))
