import concurrent.futures, json, pathlib, re, subprocess, time
out=pathlib.Path(__file__).parent
rows=json.loads((out/'readers.json').read_text())
# External test packages own these leaves.
for row in rows:
 if row['package'].endswith('/internal/ir'): row['tests']=['TestCallTargetReaders']
 if row['package'].endswith('/internal/fresh'): row['tests']=['TestFreshCorpusUnitsCoverEveryProgram']

def run(row):
 name=row['package'].removeprefix('github.com/system-inc/adamic/').replace('/','_')
 command=['go','test',row['package'],'-run','^('+'|'.join(map(re.escape,row['tests']))+')$' if row['tests'] else '^$','-count=1','-timeout','90s','-v']
 start=time.monotonic()
 try:
  with (out/('reader-'+name+'.log')).open('w') as log:
   p=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  return row|dict(command=command,exit=p.returncode,seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired:
  return row|dict(command=command,exit='build-timeout',seconds=round(time.monotonic()-start,3))
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for row in pool.map(run,rows):
  print(row['package'],row['exit'],row['seconds'],flush=True);results.append(row)
  (out/'reader-results.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(row['exit']==0 for row in results)
