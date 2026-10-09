from pathlib import Path
import subprocess,json,time,os
p=Path('review/test-defend/internal-lower-prototype')
names=['TestRegExpNativeRefusals','TestRepresentationClockSourceCheckedTypes','TestPrototypeMethodsAreRefusedWithReasons','TestSuppressionDirectiveSoundNeighbors','TestDefaultTaggedInterfaceNeedsNoFlag']
results=[]
for name in names+['rest']:
 pattern='^'+name+'$' if name!='rest' else '^('+'|'.join(x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test') and x!='TestRegExpNativeRefusals')+')$'
 argv=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/regexp','-coverprofile='+str(p/(name+'.cover')),'./internal/lower/','-run',pattern]
 env=os.environ|{'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/defend-prototype/cache/coverage'}
 start=time.monotonic()
 with (p/(name+'-coverage.log')).open('w') as f: r=subprocess.run(argv,stdout=f,stderr=subprocess.STDOUT,env=env)
 results.append({'test':name,'command':argv,'exit':r.returncode,'wall':time.monotonic()-start})
 print(name,r.returncode,results[-1]['wall'],flush=True)
 (p/'coverage-runs.json').write_text(json.dumps(results,indent=2))
 if r.returncode: break
