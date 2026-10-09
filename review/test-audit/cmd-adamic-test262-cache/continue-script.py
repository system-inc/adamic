import pathlib,subprocess,time,json,os,re
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/cmd-adamic-test262-cache';menu=json.loads((e/'menu.json').read_text());probes=json.loads((e/'probes.json').read_text());runs=json.loads((e/'mutant-runs.json').read_text());env=os.environ.copy();env['ADAMIC_TEST262_MEASURE']='1';paths={f:r/'cmd/adamic-test262'/f for f in ['cache.go','compiler.go','run.go']};bases={f:subprocess.check_output(['git','show','origin/main:cmd/adamic-test262/'+f],cwd=r).decode() for f in paths}
callers={
'PReuse':['TestNodeCacheProgram','TestNativeCacheGeneratedC','TestCacheKeyDimensions','TestCacheAtomicBytes','TestCompilerCacheProgram'],
'PObserve':['TestCacheBypass'],
'PCompilerKey':['TestCompilerCacheProgram'],
'PNodeKey':['TestNodeCacheProgram','TestCacheKeyDimensions'],
'PNativeKey':['TestNativeCacheGeneratedC','TestCacheKeyDimensions'],
'PKey':['TestCacheKeyDimensions'],
'PFilter':['TestParallelCachedMatchesSerial','TestOrderedProgress'],
'PAttempt':['TestUnavailableCacheStillRuns','TestImportedInputsRunFresh','TestCompilerCacheAcrossScratchDirectories'],
'PDependent':['TestCompilerCacheProgram'],
'PCompile':['TestCompilerCacheProgram'],
'PWorker':['TestCompilerWorkerMatchesSubprocess','TestCompilerWorkerTimeout']}
(e/'probe-rows.json').write_text(json.dumps(callers,indent=2))
def events(log):
 return [json.loads(l) for l in log.read_text().splitlines() if l.startswith('{')]
def complete(log):
 return log.exists() and any(x.get('Action') in ['pass','fail'] and 'Test' not in x for x in events(log))
def run(cmd,log,extra={}):
 start=time.monotonic()
 with (e/log).open('w') as stream:q=subprocess.run(cmd,cwd=r,env=env|extra,stdout=stream,stderr=subprocess.STDOUT)
 result=dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start,environment=extra);runs.append(result);(e/'mutant-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(result['wall'],3),flush=True);return q.returncode
# The interrupted driver's active test can finish with its already-built binary.
active=[m['id'] for m in menu if (e/(m['id']+'.log')).exists() and not complete(e/(m['id']+'.log'))]
for mid in active:
 deadline=time.monotonic()+60
 while not complete(e/(mid+'.log')) and time.monotonic()<deadline:time.sleep(.5)
 if not complete(e/(mid+'.log')):(e/(mid+'.log')).rename(e/(mid+'-interrupted.log'))
try:
 for f,p in paths.items():p.write_text((e/(f+'.switch.txt')).read_text())
 for m in menu:
  mid=m['id'];log=e/(mid+'.log')
  if complete(log):
   if not any(x['log']==mid+'.log' for x in runs):
    final=next(x for x in reversed(events(log)) if 'Test' not in x and x.get('Action') in ['pass','fail']);runs.append(dict(log=mid+'.log',command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'],exit=0 if final['Action']=='pass' else 1,wall=None,binary_seconds=final.get('Elapsed'),environment=dict(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u012/cache/'+mid),note='completed after driver interruption; wall timer unavailable'))
   continue
  run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'],mid+'.log',dict(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u012/cache/'+mid))
 for probe in probes:
  mid=probe['id'];pattern='^('+'|'.join(callers[mid])+')$'
  run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run',pattern],mid+'.log',dict(ADAMIC_MUTANT=mid,ADAMIC_BUILD_CACHE_DIR='/tmp/u012/cache/'+mid))
finally:
 for f,p in paths.items():p.write_text(bases[f])
 (e/'mutant-runs.json').write_text(json.dumps(runs,indent=2))
