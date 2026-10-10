import json,subprocess,time
from pathlib import Path
ev=Path(__file__).resolve().parent;root=ev.parents[2];paths=json.loads((ev/'admission-inputs.json').read_text());rows=[]
for i in range(0,len(paths),64):
 chunk=paths[i:i+64];pair=[]
 for name,binary in [('main','/tmp/chain-slice-3-main-admission'),('slice','/tmp/chain-slice-3-admission')]:
  cmd=['timeout','85',binary]+chunk;log=ev/('admission-%s-%03d.jsonl'%(name,i//64));begin=time.monotonic()
  with log.open('w') as out:r=subprocess.run(cmd,cwd=root,stdout=out,stderr=subprocess.STDOUT)
  assert r.returncode==0,(name,i,r.returncode)
  records=[json.loads(s) for s in log.read_text().splitlines()];assert len(records)==len(chunk)
  pair.append(records);rows.append(dict(mode=name,shard=i//64,exit=r.returncode,seconds=time.monotonic()-begin,inputs=len(chunk)))
 print(i//64,len(chunk),flush=True)
 (ev/'admission-shards.json').write_text(json.dumps(rows,indent=2)+'\n')
main={};final={}
for mode,dest in [('main',main),('slice',final)]:
 for log in sorted(ev.glob('admission-'+mode+'-*.jsonl')):
  for s in log.read_text().splitlines():
   r=json.loads(s);dest[r['Path']]=r
changes=[dict(path=p,before=main[p],after=final[p]) for p in paths if main[p]!=final[p]]
(ev/'admission-delta.json').write_text(json.dumps(dict(base='5e33a17b186a8a2218d27b69b21e2de5acc5b750',inputs=len(paths),changed=changes),indent=2)+'\n');print('DELTA',len(changes),flush=True)
