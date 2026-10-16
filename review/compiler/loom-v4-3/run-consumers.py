import concurrent.futures,json,re,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4-3';packages=json.loads((e/'consumers.json').read_text());selection={};actual={}
def prepare(p):
 package=p['package'];binary=e/(package.replace('/','-')+'-consumer.test');start=time.monotonic()
 with (e/(package.replace('/','-')+'-consumer-build.log')).open('w') as log:
  try:r=subprocess.run(['go','test','-c','-o',str(binary),'./'+package],stdout=log,stderr=subprocess.STDOUT,timeout=180);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 row=dict(package=package,build=True,seconds=round(time.monotonic()-start,3),exit=code)
 if code or not binary.exists():row['no_tests']=not binary.exists();return row
 names=set(x for x in subprocess.check_output([str(binary),'-test.list','^Test'],cwd=root/package,text=True,timeout=30).splitlines() if x.startswith('Test'));actual[package]=names
 ordinary=[]
 for f in p['test_files']:
  if any(token in f for token in ['generated','shards','registered_units']):continue
  ordinary+=re.findall(r'^func (Test\w+)\(t \*testing.T\)',Path(package,f).read_text(),re.M)
 ordinary=[t for t in ordinary if t in names and not t.startswith('TestProduct_') and not any(token in t for token in ['Throughput','Timing','Benchmark','Setup','Union','Planted'])]
 # Prefer behavioral checks that exercise a compiler-backed port, then bounded gaps/invariants.
 def rank(t):return (not any(x in t for x in ['AgreesWith','MatchesGo','ParsesAsGo','MakesWhatGo','CompileProfiles','WhatTheCompilerLinks','ThrowingAccessorSpread']),t)
 tests=sorted(set(ordinary),key=rank)[:3]
 if not tests:tests=sorted(t for t in names if not t.startswith('TestProduct_') and not any(x in t for x in ['Union','Setup','Throughput','Timing']))[:3]
 products=sorted(t for t in names if t.startswith('TestProduct_') and not any(x in t for x in ['WASI','Wasm','TSan','Pinned','Mutant','Corpus']))
 if len(products)>3:
  def productrank(t):return (not any(x in t.lower() for x in ['lowered','source','oracle']),t)
  products=sorted(products,key=productrank)[:3]
 if package=='internal/native':tests=['TestConstantPortabilityClang','TestLongUnicodePortabilityClang','TestLongUnicodeLiteralByteCoverage']
 selection[package]=dict(products=products,tests=tests);return row
def run(package,test):
 start=time.monotonic();path=e/(package.replace('/','-')+'-'+test+'-consumer.log')
 with path.open('w') as log:
  try:r=subprocess.run([str(e/(package.replace('/','-')+'-consumer.test')),'-test.run','^'+test+'$','-test.count=1','-test.timeout=90s','-test.v','-test.parallel=1'],cwd=root/package,stdout=log,stderr=subprocess.STDOUT,timeout=95);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 seconds=round(time.monotonic()-start,3);output=path.read_text(errors='replace');row=dict(package=package,test=test,seconds=seconds,exit=code,grain=seconds<60,skip=('--- SKIP: '+test) in output);print(row,flush=True);return row
def package_checks(package):
 rows=[]
 for test in selection[package]['products']:
  row=run(package,test);rows.append(row)
  if row['exit'] or not row['grain']:
   rows.append(dict(package=package,blocked='setup product failed or exceeded its grain; remaining package checks not run'));return rows
 for test in selection[package]['tests']:
  if test in actual[package]:rows.append(run(package,test))
 return rows
with (e/'consumer-results.jsonl').open('w') as out:
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  for row in pool.map(prepare,packages):out.write(json.dumps(row)+'\n');out.flush();print(row,flush=True)
 (e/'selected-consumer-tests.json').write_text(json.dumps(selection,indent=2)+'\n')
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  for rows in pool.map(package_checks,selection):
   for row in rows:out.write(json.dumps(row)+'\n')
   out.flush()
for p in packages:
 binary=e/(p['package'].replace('/','-')+'-consumer.test')
 if binary.exists():binary.unlink()
