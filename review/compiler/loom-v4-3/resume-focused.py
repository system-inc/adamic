import concurrent.futures,json,re,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4-3'
selected=json.loads((e/'selected-focused.json').read_text())
records=[json.loads(s) for s in (e/'focused-results.jsonl').read_text().splitlines()]
completed={(r['package'],r['test']) for r in records if 'test' in r}
recovered=[];jobs=[]
for package,tests in selected.items():
 for test in tests:
  if (package,test) in completed:continue
  path=e/(package.replace('/','-')+'-'+test+'.log')
  output=path.read_text() if path.exists() else ''
  passed=re.search(r'^--- (PASS|SKIP): '+re.escape(test)+r' \(([0-9.]+)s\)',output,re.M)
  if passed:
   seconds=float(passed[2]);recovered.append(dict(package=package,test=test,seconds=seconds,exit=0,grain=seconds<60,skip=passed[1]=='SKIP',timing='test leaf from completed log; driver interrupted before recording wall time'))
  elif '--- FAIL: '+test in output:
   raise RuntimeError('Unrecorded failing test is not rerun: '+test)
  else:jobs.append((package,test))
def run(job):
 package,test=job;path=e/(package.replace('/','-')+'-'+test+'.log');start=time.monotonic()
 with path.open('w') as log:
  try:r=subprocess.run([str(e/(package.replace('/','-')+'.test')),'-test.run','^'+test+'$','-test.count=1','-test.timeout=90s','-test.parallel=1','-test.v'],cwd=root/package,stdout=log,stderr=subprocess.STDOUT,timeout=95);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 seconds=round(time.monotonic()-start,3);output=path.read_text();row=dict(package=package,test=test,seconds=seconds,exit=code,grain=seconds<60,skip=('--- SKIP: '+test) in output);print(row,flush=True);return row
with (e/'focused-results.jsonl').open('a') as out:
 for row in recovered:out.write(json.dumps(row)+'\n');out.flush()
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  for row in pool.map(run,jobs):out.write(json.dumps(row)+'\n');out.flush()
print('Resumed',len(jobs),'uncompleted tests; recovered',len(recovered),'completed logs',flush=True)
