import concurrent.futures,json,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-v4-3/fixed/readers';selected=json.loads((root/'review/compiler/loom-r-v4/reader-selection.json').read_text())
def check(package):
 start=time.monotonic();binary=e/(package.replace('/','-')+'-reader.test')
 with (e/(package.replace('/','-')+'-reader-build.log')).open('w') as log:
  try:r=subprocess.run(['go','test','-c','-o',str(binary),'./'+package],stdout=log,stderr=subprocess.STDOUT,timeout=180);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 rows=[dict(package=package,build=True,seconds=round(time.monotonic()-start,3),exit=code)]
 if code:return rows
 if not binary.exists():rows[-1]['no_tests']=True;return rows
 actual=set(subprocess.check_output([str(binary),'-test.list','^Test'],cwd=root/package,text=True,timeout=30).splitlines())
 for test in selected[package]:
  if test not in actual:raise RuntimeError('missing '+package+' '+test)
  start=time.monotonic();path=e/(package.replace('/','-')+'-'+test+'-reader.log')
  with path.open('w') as log:
   try:r=subprocess.run([str(binary),'-test.run','^'+test+'$','-test.count=1','-test.timeout=90s','-test.parallel=1','-test.v'],cwd=root/package,stdout=log,stderr=subprocess.STDOUT,timeout=95);code=r.returncode
   except subprocess.TimeoutExpired:code=124
  seconds=round(time.monotonic()-start,3);text=path.read_text();rows.append(dict(package=package,test=test,seconds=seconds,exit=code,grain=seconds<60,skip=('--- SKIP: '+test) in text));print(rows[-1],flush=True)
 return rows
with (e/'reader-results.jsonl').open('w') as out:
 with concurrent.futures.ThreadPoolExecutor(max_workers=1) as pool:
  for rows in pool.map(check,selected):
   for row in rows:out.write(json.dumps(row)+'\n');print(row,flush=True)
   out.flush()

(e/"done").write_text("finished\n")
