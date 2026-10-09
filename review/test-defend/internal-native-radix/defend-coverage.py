import os,pathlib,subprocess,json,time,concurrent.futures
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');os.chdir('/workspace/adamic');env=dict(os.environ,ADAMIC_RECORD_BENCH='1')
rows=['TestToStringWithARadixOutOfRangePanics','TestToStringWithARadixMatchesNode','TestRecordBenchmark','TestRuntimeStringEquality','TestRegExpSearchNode','TestRegExpLintPatternsNode','TestRegExpBytecodeTest262','TestRegExpNativeStepLimit','TestRegExpBytecodePatternUnits','TestRuntimeReleasePaths','TestRegExpIteratorResultShape','TestRegExpBytecodeRandomNode family']
def run(n):
 pat='^TestRegExpBytecodeRandomNodeUnit[0-9]+$' if n.endswith('family') else '^'+n+'$';key=n.replace(' ','_');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/native,./internal/regexp','-coverprofile='+str(p/(key+'.cover')),'./internal/native/','-run',pat]
 t=time.monotonic()
 with (p/(key+'-coverage.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 print(n,r.returncode,round(time.monotonic()-t,2),flush=True);return {'row':n,'exit':r.returncode,'command':cmd}
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:a=list(pool.map(run,rows))
(p/'coverage-runs.json').write_text(json.dumps(a,indent=2))
