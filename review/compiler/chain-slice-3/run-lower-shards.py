import json,subprocess,time
from pathlib import Path
ev=Path(__file__).resolve().parent;root=ev.parents[2];start=time.monotonic()
with (ev/'lower-list.log').open('w') as out:r=subprocess.run(['timeout','90','go','test','./internal/lower','-list','^Test','-timeout','85s'],cwd=root,stdout=out,stderr=subprocess.STDOUT)
assert r.returncode==0
names=[s for s in (ev/'lower-list.log').read_text().splitlines() if s.startswith('Test')];rows=[]
for i in range(0,len(names),32):
 selected=names[i:i+32];cmd=['timeout','90','go','test','./internal/lower','-run','^('+'|'.join(selected)+')$','-count=1','-timeout','85s','-parallel','4','-json'];begin=time.monotonic();log=ev/('lower-shard-%03d.jsonl'%(i//32))
 with log.open('w') as out:r=subprocess.run(cmd,cwd=root,stdout=out,stderr=subprocess.STDOUT)
 row=dict(shard=i//32,tests=selected,exit=r.returncode,seconds=time.monotonic()-begin,command=cmd);rows.append(row);(ev/'lower-shards.json').write_text(json.dumps(rows,indent=2)+'\n');print(row['shard'],row['exit'],round(row['seconds'],3),flush=True)
 if r.returncode:raise RuntimeError(str(log))
print('PASS',len(names),'tests',len(rows),'shards',flush=True)
