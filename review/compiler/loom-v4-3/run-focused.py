import concurrent.futures,json,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-v4-3';selected=json.loads((root/'review/compiler/loom-v4/selected-tests.json').read_text())
selected['internal/oracle']+=['TestProduct_V4SanitizedCountedRuntimes'+str(i) for i in range(3)]+['TestRuntimeSetsAreEachUsed','TestRuntimeProductKeyHoldsEveryPart','TestFamilyWrappersAreCurrent']
selected['internal/ir']+=['TestObjectSpreadAccessorEffects', 'TestEveryMapAProgramHoldsIsSorted', 'TestCallTargetReaders']
selected['internal/lower']+=['TestNativeAgreementRowsNameRegisteredPrograms', 'TestClassFeaturesThrowingAccessorSpread','TestClassFeaturesNarrowedAccessor','TestClassFeaturesAccessorRefusals']
selected['internal/native']+=['TestConstantBytesAreUnsigned','TestGeneratedCIsPortable']
# Main's new fixtures are reached through their hash-owned native oracle shards.
for fixture in ['accessor_spread_throw_first','accessor_spread_throw_middle','accessor_spread_throw_last','plain_data_spread','template_nullish','static_constructor_spread_gap','static_constructor_spread_owned']:
 key='internal/oracle/testdata/'+fixture+'.a';h=14695981039346656037
 for b in key.encode():h=((h^b)*1099511628211)&((1<<64)-1)
 selected['internal/oracle'].append('TestNativeAgreesWithNode_'+str(h%128).zfill(3))
# List actual test names rather than silently accepting an absent selector.
def build(package):
 binary=e/(package.replace('/','-')+'.test')
 with (e/(package.replace('/','-')+'-build.log')).open('w') as log:
  start=time.monotonic();r=subprocess.run(['go','test','-c','-o',str(binary),'./'+package],stdout=log,stderr=subprocess.STDOUT,timeout=180)
 if r.returncode:raise RuntimeError(package+' build failed')
 listing=subprocess.check_output([str(binary),'-test.list','^Test'],cwd=root/package,text=True,timeout=30).splitlines();actual=set(x for x in listing if x.startswith('Test'))
 missing=set(selected[package])-actual
 if missing:
  print('absent requested tests',package,sorted(missing),flush=True);selected[package]=[x for x in selected[package] if x in actual]
 return dict(package=package,build=True,seconds=round(time.monotonic()-start,3),exit=0)
def check(job):
 p,t=job;start=time.monotonic();logpath=e/(p.replace('/','-')+'-'+t+'.log')
 with logpath.open('w') as log:
  try:r=subprocess.run([str(e/(p.replace('/','-')+'.test')),'-test.run','^'+t+'$','-test.count=1','-test.timeout=90s','-test.v','-test.parallel=1'],cwd=root/p,stdout=log,stderr=subprocess.STDOUT,timeout=95);code=r.returncode
  except subprocess.TimeoutExpired:code=124
 seconds=round(time.monotonic()-start,3);output=logpath.read_text(errors='replace');row=dict(package=p,test=t,seconds=seconds,exit=code,grain=seconds<60,skip=('--- SKIP: '+t) in output);print(row,flush=True);return row
with (e/'focused-results.jsonl').open('w') as out:
 for package in selected:
  row=build(package);out.write(json.dumps(row)+'\n');out.flush()
 selected={p:sorted(set(tests)) for p,tests in selected.items()};(e/'selected-focused.json').write_text(json.dumps(selected,indent=2)+'\n')
 for p,tests in selected.items():
  for t in tests:
   if t.startswith('TestProduct_'):
    row=check((p,t));out.write(json.dumps(row)+'\n');out.flush()
 jobs=[(p,t) for p,tests in selected.items() for t in tests if not t.startswith('TestProduct_')]
 with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  for row in pool.map(check,jobs):out.write(json.dumps(row)+'\n');out.flush()
for p in selected:(e/(p.replace('/','-')+'.test')).unlink()
