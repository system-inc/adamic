import concurrent.futures,json,subprocess,time
from pathlib import Path
e=Path('review/compiler/loom-r-unions-ac');standard=['-count=1','-timeout','90s','-parallel','4','-json'];cases=[
('oracle-path-readers',['go','test','./internal/oracle','-run','^Test(NullableObjectStorage.*|NullableStringFieldCheckMutant|NonNullAdamicRefusal|CountsAreRecordedTable)$',*standard],180),
('union-fixtures',['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/scout_',*standard],180),
('reader-guard',['go','test','./internal/ir','-run','^TestCallTargetReaders$',*standard],180),
('vet',['go','vet','./internal/lower','./internal/native','./internal/oracle','./stage1/cohere/selector','./stage1/cohere/values','./internal/ir'],180)]
def run(c):
 label,cmd,limit=c;start=time.monotonic()
 with (e/(label+'.log')).open('w') as f:
  try:code=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired:code=124
 events=[]
 for l in (e/(label+'.log')).read_text().splitlines():
  try:r=json.loads(l)
  except:continue
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:events.append(r)
 row=dict(label=label,command=cmd,seconds=round(time.monotonic()-start,3),exit=code,tests=events);print(label,code,row['seconds'],flush=True);return row
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:rows=list(pool.map(run,cases))
(e/'check-results.json').write_text(json.dumps(rows,indent=2)+'\n');assert all(r['exit']==0 for r in rows);assert all(v.get('Elapsed',0)<60 for r in rows for v in r['tests'])
