"""Include runtime-library setup in each isolated test invocation."""
import json, os, shutil, subprocess, time
from pathlib import Path
root=Path(__file__).resolve().parents[3]; out=Path(__file__).resolve().parent
base=dict(os.environ,GOMAXPROCS='4',ADAMIC_GATE_UNCACHED='1')
base['GOCACHE']=subprocess.check_output(['go','env','GOCACHE'],text=True).strip()
rows=[]
for pkg,test in [('./internal/ir','TestCallMayThrowUsesReachableTargets'),('./internal/ir','TestDirectClosureTargetsUseEncodedIndex'),('./internal/oracle','TestCallTargetThrowAgreesWithNode'),('./internal/oracle','TestDirectClosureCallAgreesWithNode')]:
 cache=Path('/workspace/call-targets-cold')/test
 if cache.exists(): shutil.rmtree(cache)
 env=dict(base,XDG_CACHE_HOME=str(cache))
 command=['go','test',pkg,'-run','^'+test+'$','-count=1','-json','-timeout=90s']
 log=out/('cold-'+test+'.jsonl'); start=time.monotonic()
 with log.open('w') as stream: result=subprocess.run(command,cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT)
 seconds=time.monotonic()-start
 events=[]
 for line in log.read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError: pass
 terminal=[e for e in events if e.get('Test')==test and e.get('Action') in ['pass','fail','skip']]
 row=dict(test=test,command=command,invocation_seconds=round(seconds,3),test_seconds=terminal[-1].get('Elapsed') if terminal else None,action=terminal[-1]['Action'] if terminal else 'build-failure',exit=result.returncode)
 rows.append(row); (out/'cold-results.json').write_text(json.dumps(rows,indent=2)+'\n'); print(row,flush=True)
 if cache.exists(): shutil.rmtree(cache)
 assert row['action']=='pass' and result.returncode==0 and seconds<60,(row,log.read_text()[-5000:])
