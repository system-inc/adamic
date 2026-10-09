from pathlib import Path
import subprocess,json,time,re
p=Path('review/test-audit/bridge-tsgo-product_units');source=Path('bridge/tsgo/product_units_test.go').read_text();rows=re.findall(r'func (TestProduct_\w+)\(',source)
(p/'rows.json').write_text(json.dumps(rows,indent=2));records=[]
for row in rows:
 for trial in range(1,4):
  cmd=['timeout','120','go','test','-count=1','-timeout','90s','./bridge/tsgo/','-run','^'+row+'$'];start=time.monotonic()
  path=p/(row+f'-{trial}.log')
  with path.open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  match=re.search(r'\t([0-9.]+)s',path.read_text());record={'test':row,'trial':trial,'exit':r.returncode,'seconds':float(match[1]) if match else None,'wall':time.monotonic()-start,'command':cmd};records.append(record)
  (p/'timings.json').write_text(json.dumps(records,indent=2))
  if r.returncode:print(record,flush=True);raise SystemExit('red isolated run')
 print(row, [x['seconds'] for x in records if x['test']==row],flush=True)
