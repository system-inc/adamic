import pathlib,subprocess,time,json,os,re
r=pathlib.Path('/workspace/adamic');e=r/'review/test-audit/cmd-adamic-test262-cache'
rows='''TestNodeCacheProgram TestNativeCacheGeneratedC TestCacheKeyDimensions TestCacheAtomicBytes TestCacheBypass TestParallelCachedMatchesSerial TestOrderedProgress TestUnavailableCacheStillRuns TestImportedInputsRunFresh TestCompilerCacheProgram TestCompilerCacheAcrossScratchDirectories TestCompilerWorkerMatchesSubprocess TestCompilerHangHelper TestCompilerWorkerTimeout'''.split()
(e/'requested-rows.json').write_text(json.dumps(rows,indent=2));env=os.environ.copy();env['ADAMIC_TEST262_MEASURE']='1';runs=[]
def run(cmd,log):
 start=time.monotonic()
 with (e/log).open('w') as f: q=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start));(e/'clean-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,flush=True);return q.returncode
pattern='^('+'|'.join(rows)+')$'
run(['timeout','120','go','test','-count=1','-timeout','90s','-coverprofile='+str(e/'slice.cover'),'./cmd/adamic-test262/','-run',pattern],'slice-coverage.log')
run(['go','tool','cover','-func='+str(e/'slice.cover')],'functions-coverage.txt')
for row in rows:
 for n in range(1,4):
  if run(['timeout','120','go','test','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','^'+row+'$'],f'timing-{row}-{n}.log')!=0:raise SystemExit('clean row failed: '+row)
