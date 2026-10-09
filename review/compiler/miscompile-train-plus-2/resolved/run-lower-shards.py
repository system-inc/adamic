from pathlib import Path
import subprocess,json,time
out=Path(__file__).parent
names=[s for s in (out/'lower-list.log').read_text().splitlines() if s.startswith('Test')]
assert names and len(names)==len(set(names))
shards=[names[i::4] for i in range(4)]
assert sorted(sum(shards,[]))==sorted(names)
(out/'lower-shards.json').write_text(json.dumps(shards,indent=2)+'\n')
results=[]
for i,names in enumerate(shards):
 command=['go','test','./internal/lower','-run','^('+'|'.join(names)+')$','-count=1','-v','-timeout','90s']
 start=time.monotonic()
 with (out/('lower-shard-'+str(i+1)+'.log')).open('w') as log:
  r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=120)
 results.append({'shard':i+1,'tests':len(names),'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'command':command})
 (out/'lower-results.json').write_text(json.dumps(results,indent=2)+'\n')
 print({k:v for k,v in results[-1].items() if k!='command'},flush=True)
 assert r.returncode==0
