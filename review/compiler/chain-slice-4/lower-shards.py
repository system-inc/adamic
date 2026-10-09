import json,os,re,subprocess,time
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/chain-slice-4';names=sorted(set(n for p in (root/'internal/lower').glob('*_test.go') for n in re.findall(r'^func (Test\w+)\(',p.read_text(),re.M)))
rows=[]
for i in range(0,len(names),25):
 selected=names[i:i+25];cmd=['go','test','./internal/lower','-run','^('+'|'.join(selected)+')$','-count=1','-timeout','85s','-v'];start=time.monotonic()
 with (out/f'lower-shard-{i//25}.log').open('w') as f:
  try:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=90);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 rows.append({'shard':i//25,'names':selected,'exit':code,'seconds':round(time.monotonic()-start,3)})
 print(rows[-1]['shard'],code,rows[-1]['seconds'],flush=True)
(out/'lower-shards.json').write_text(json.dumps(rows,indent=2)+'\n')
if any(r['exit'] for r in rows):raise SystemExit(1)
