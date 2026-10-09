import pathlib,json,subprocess,time,concurrent.futures,os
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-css-composition_shards';groups=json.loads((p/'row-members.json').read_text());runs=[]
def complete(f):
 if not f.exists():return False
 for l in f.read_text().splitlines():
  try:x=json.loads(l)
  except:continue
  if x.get('Action')=='pass' and 'Test' not in x:return True
 return False
def rowrun(item):
 row,ts=item;rr=[]
 for i in range(1,4):
  log='timing-'+row.replace(' ','_')+'-'+str(i)+'.log'
  if complete(p/log):continue
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^('+'|'.join(ts)+')$'];start=time.monotonic()
  with (p/log).open('w') as f:q=subprocess.run(cmd,cwd=r,stdout=f,stderr=subprocess.STDOUT)
  rr.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));print(log,q.returncode,round(rr[-1]['wall'],3),flush=True)
  if q.returncode!=0:break
 return rr
# First schedule independent remaining rows with two workers; family measurements run alone.
items=[(row,ts) for row,ts in groups.items() if ' family' not in row]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as ex:
 for result in ex.map(rowrun,items):runs.extend(result);(p/'timing-runs.json').write_text(json.dumps(runs,indent=2))
# Do not touch the currently completing composition-family measurement.
for row,ts in groups.items():
 if row=='TestCSSPrinterAgreesWithGo family':runs.extend(rowrun((row,ts)));(p/'timing-runs.json').write_text(json.dumps(runs,indent=2))
print('timing phase complete',flush=True)
