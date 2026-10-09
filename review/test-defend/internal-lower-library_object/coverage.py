from pathlib import Path
import subprocess,json,time,os
p=Path('review/test-defend/internal-lower-library_object')
names=['TestLibraryStringRefusals','TestWhatZeroOneRefusesIsRefusedWithAFix','TestATupleSeenAsAnArrayIsNotYet','TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat','TestAMethodReadAsAValueIsRefused']
results=[]
for name in names:
 argv=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower','-coverprofile='+str(p/(name+'.cover')),'./internal/lower/','-run','^'+name+'$']
 start=time.monotonic()
 with (p/(name+'-coverage.log')).open('w') as f:r=subprocess.run(argv,stdout=f,stderr=subprocess.STDOUT,env=os.environ|{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-library-object/cache/coverage'})
 results.append({'test':name,'command':argv,'exit':r.returncode,'wall':time.monotonic()-start});print(name,r.returncode,results[-1]['wall'],flush=True)
 (p/'coverage-runs.json').write_text(json.dumps(results,indent=2))
 if r.returncode:break
